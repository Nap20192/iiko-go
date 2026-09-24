package corporation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
	"github.com/Nap20192/iiko-go/iikoserver/rest/resttest"
)

// newFake wires a Service to a stub iiko; the transport is returned too, for the
// tests that close the session.
func newFake(t *testing.T, h http.HandlerFunc) (*Service, *rest.Client) {
	t.Helper()
	c := resttest.NewFake(t, h)
	return New(c), c
}

func TestAccountsListFallsBackToOtherDocumentedPath(t *testing.T) {
	// The docs print the path two ways on two independent pages. A build that
	// only honours the header spelling must still work.
	var tried []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/resto/api/auth" {
			_, _ = w.Write([]byte("tok"))
			return
		}
		tried = append(tried, r.URL.Path)
		if strings.Contains(r.URL.Path, "/v2/") {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
			return
		}
		_, _ = w.Write([]byte(`[{"id":"a1","name":"Касса","type":"CASH","code":"1.01"}]`))
	}))
	defer srv.Close()

	s := New(rest.New(rest.Config{BaseURL: srv.URL + "/resto", Login: "u", Password: "p"}))
	ents, err := s.ListEntities(context.Background(), KindAccounts, false)
	if err != nil {
		t.Fatalf("fallback did not happen: %v", err)
	}
	if len(ents) != 1 || ents[0].Name != "Касса" {
		t.Fatalf("bad decode after fallback: %+v", ents)
	}
	if len(tried) != 2 || !strings.Contains(tried[0], "/v2/") || strings.Contains(tried[1], "/v2/") {
		t.Fatalf("expected /v2/ first then the header spelling, got %v", tried)
	}
}
func TestAccountsListDoesNotFallBackOnRealErrors(t *testing.T) {
	// A 403 means the path was right and the rights were wrong — retrying the
	// other spelling would just burn a request and mask the real problem.
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/resto/api/auth" {
			_, _ = w.Write([]byte("tok"))
			return
		}
		calls++
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Permission denied"))
	}))
	defer srv.Close()

	s := New(rest.New(rest.Config{BaseURL: srv.URL + "/resto", Login: "u", Password: "p"}))
	if _, err := s.ListEntities(context.Background(), KindAccounts, false); err == nil {
		t.Fatal("expected the 403 to surface")
	}
	if calls != 1 {
		t.Fatalf("403 must not trigger the path fallback, got %d calls", calls)
	}
}

func TestServerTypeUnquotesTheBareResponse(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, body, want string }{
		{"bare token", "CHAIN", "CHAIN"},
		{"quoted", `"STANDALONE_RMS"`, "STANDALONE_RMS"},
		{"with trailing newline", "REPLICATED_RMS\n", "REPLICATED_RMS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(tt.body)) })
			got, err := s.ServerType(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEntityKindsMatchesTheSwitch(t *testing.T) {
	t.Parallel()
	// The tool enum is built from EntityKinds, so a kind listed here but not
	// handled by ListEntities would be offered to the model and then rejected.
	for _, k := range EntityKinds() {
		s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.Path, "accounts"):
				_, _ = w.Write([]byte(`[]`)) // the only JSON one
			case strings.Contains(r.URL.Path, "suppliers"):
				// Suppliers are Users with supplier=true, so this is employee XML
				// despite being a "list" endpoint.
				_, _ = w.Write([]byte(`<employees></employees>`))
			default:
				_, _ = w.Write([]byte(`<corporateItemDtoes></corporateItemDtoes>`))
			}
		})
		if _, err := s.ListEntities(context.Background(), EntityKind(k), false); err != nil {
			t.Errorf("kind %q is advertised but not handled: %v", k, err)
		}
	}
	if _, err := New(resttest.NewFake(t, func(w http.ResponseWriter, r *http.Request) {})).
		ListEntities(context.Background(), EntityKind("nonsense"), false); err == nil {
		t.Error("an unknown kind must be rejected")
	}
}

// Suppliers moved to their own domain: they are Users with supplier=true and
// share nothing with the corporation structure but a response format. Leaving a
// silently-working alias here would keep two ways to reach one endpoint.
func TestListEntitiesNoLongerServesSuppliers(t *testing.T) {
	t.Parallel()
	s, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("suppliers must not be fetched through corporation any more")
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := s.ListEntities(context.Background(), KindSuppliers, false)
	if err == nil {
		t.Fatal("KindSuppliers must be rejected here")
	}
	if !strings.Contains(err.Error(), "suppliers") {
		t.Errorf("the error must point at the suppliers domain, got %v", err)
	}
	for _, k := range EntityKinds() {
		if k == "suppliers" {
			t.Error("EntityKinds must no longer advertise suppliers")
		}
	}
}
