package rest

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newFake wires a Client to a stub iiko that always authenticates, so a test
// only has to describe the endpoint it actually cares about.
// departments is a minimal decode target: these tests exercise the transport,
// not the corporation domain, so they must not depend on its types.
type departments struct {
	Items []struct {
		ID   string `xml:"id"`
		Name string `xml:"name"`
	} `xml:"corporateItemDto"`
}

const pathDepartments = "/api/corporation/departments"

func newFake(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/auth") {
			_, _ = w.Write([]byte("tok"))
			return
		}
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	return New(Config{BaseURL: srv.URL + "/resto", Login: "u", Password: "p"})
}

func TestClient_AllowWrite(t *testing.T) {
	t.Parallel()
	if New(Config{}).AllowWrite() {
		t.Error("writes must be off unless configured")
	}
	if !New(Config{AllowWrite: true}).AllowWrite() {
		t.Error("AllowWrite must reflect config")
	}
}

func TestReauthOn401AndLogoutReleasesSlot(t *testing.T) {
	var auths, logouts atomic.Int32
	live := true

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/resto/api/auth":
			auths.Add(1)
			live = true
			_, _ = w.Write([]byte("token-" + strings.Repeat("a", 8)))
		case "/resto/api/logout":
			logouts.Add(1)
			_, _ = w.Write([]byte(r.URL.Query().Get("key")))
		case "/resto/api/corporation/departments":
			if !live {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("session expired"))
				return
			}
			_, _ = w.Write([]byte(`<corporateItemDtoes><corporateItemDto><id>d1</id><name>Тверская</name><type>DEPARTMENT</type></corporateItemDto></corporateItemDtoes>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL + "/resto", Login: "u", Password: "p"})
	ctx := context.Background()

	var out departments
	if err := c.GetXML(ctx, pathDepartments, nil, &out); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if got := auths.Load(); got != 1 {
		t.Fatalf("expected exactly 1 auth (one licence slot), got %d", got)
	}

	// Simulate the ~1h idle timeout: next call must re-auth once and succeed.
	live = false
	out = departments{}
	if err := c.GetXML(ctx, pathDepartments, nil, &out); err != nil {
		t.Fatalf("after expiry: %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].Name != "Тверская" {
		t.Fatalf("bad XML decode: %+v", out.Items)
	}
	if got := auths.Load(); got != 2 {
		t.Fatalf("expected re-auth exactly once, got %d auths", got)
	}

	if err := c.Close(ctx); err != nil {
		t.Fatalf("close: %v", err)
	}
	if logouts.Load() != 1 {
		t.Fatal("Close must call logout — a leaked token holds a licence seat and can lock a cashier out of iikoOffice")
	}
	if err := c.Close(ctx); err != nil || logouts.Load() != 1 {
		t.Fatal("second Close must be a no-op")
	}
}
func TestRequestsAreSerialized(t *testing.T) {
	var inFlight, maxSeen atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/resto/api/auth" {
			_, _ = w.Write([]byte("tok"))
			return
		}
		n := inFlight.Add(1)
		for {
			old := maxSeen.Load()
			if n <= old || maxSeen.CompareAndSwap(old, n) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		inFlight.Add(-1)
		_, _ = w.Write([]byte(`<corporateItemDtoes></corporateItemDtoes>`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL + "/resto", Login: "u", Password: "p"})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var out departments
			_ = c.GetXML(context.Background(), pathDepartments, nil, &out)
		}()
	}
	wg.Wait()

	// iiko docs: "запросы должны выполняться последовательно".
	if got := maxSeen.Load(); got != 1 {
		t.Fatalf("saw %d concurrent requests; iiko requires strictly sequential calls", got)
	}
}
func TestLicenceExhaustionGetsActionableHint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("License enhancement is required: no connections available for module 2200"))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL + "/resto", Login: "u", Password: "p"})
	var out departments
	err := c.GetXML(context.Background(), pathDepartments, nil, &out)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "licence slot") && !strings.Contains(err.Error(), "free licence slots") {
		t.Fatalf("error must name the recovery action, got: %v", err)
	}
}

func TestLicenceInfoPassesModuleID(t *testing.T) {
	t.Parallel()
	var gotQ url.Values
	c := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQ = r.URL.Query()
		_, _ = w.Write([]byte("2 of 5 connections free"))
	})
	got, err := c.LicenceInfo(context.Background(), "2200")
	if err != nil {
		t.Fatal(err)
	}
	if gotQ.Get("moduleId") != "2200" {
		t.Errorf("moduleId = %q, want 2200", gotQ.Get("moduleId"))
	}
	if got != "2 of 5 connections free" {
		t.Errorf("body should pass through verbatim, got %q", got)
	}
}

// PUT is iiko's full-replace verb; the body must go out as XML under the right
// method, or a replace silently becomes something else.
func TestPutXMLSendsAnXMLBodyWithPUT(t *testing.T) {
	t.Parallel()
	type doc struct {
		XMLName xml.Name `xml:"employee"`
		Name    string   `xml:"name"`
	}
	var gotMethod, gotBody, gotCT string
	c := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotMethod, gotBody, gotCT = r.Method, string(b), r.Header.Get("Content-Type")
		_, _ = w.Write([]byte(`<employee><name>stored</name></employee>`))
	})
	defer c.Close(context.Background())

	var out doc
	if err := c.PutXML(context.Background(), "/api/employees/byId/e-1", nil, doc{Name: "sent"}, &out); err != nil {
		t.Fatalf("put: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if !strings.Contains(gotCT, "xml") {
		t.Errorf("Content-Type = %q, want XML", gotCT)
	}
	if !strings.Contains(gotBody, "<name>sent</name>") {
		t.Errorf("body must carry the entity, got %q", gotBody)
	}
	if out.Name != "stored" {
		t.Errorf("response must decode, got %q", out.Name)
	}
}

// A write with nothing to decode still has to report a transport or business
// failure; only the body is skipped.
func TestPostV2WithNoOutputStillSurfacesErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr bool
	}{
		{name: "success with an ignored body", status: 200, body: `{"result":"SUCCESS"}`},
		{name: "empty body is fine when nothing is decoded", status: 200, body: ``},
		{name: "transport failure still surfaces", status: 409, body: "уже проведена", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			})
			defer c.Close(context.Background())
			_, err := c.PostV2(context.Background(), "/api/v2/x", nil, map[string]string{"a": "b"}, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
