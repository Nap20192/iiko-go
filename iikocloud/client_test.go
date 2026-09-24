package iikocloud_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/Nap20192/iiko-go/iikocloud"
	"github.com/Nap20192/iiko-go/iikocloud/gen"
)

// tokenPath is served by rest, not by a domain, so it is not a domain's to cover.
const tokenPath = "/api/v2/access_token"

type recorder struct {
	mu    sync.Mutex
	paths []string
}

func (r *recorder) last() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.paths) == 0 {
		return ""
	}
	return r.paths[len(r.paths)-1]
}

// dialFake answers every call with JSON null, which decodes into a struct, a
// slice or nothing at all, so one handler serves all four method shapes.
func dialFake(t *testing.T) (*iikocloud.Client, *recorder) {
	t.Helper()
	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == tokenPath {
			json.NewEncoder(w).Encode(map[string]string{"token": "T", "correlationId": "c"})
			return
		}
		rec.mu.Lock()
		rec.paths = append(rec.paths, r.URL.Path)
		rec.mu.Unlock()
		w.Write([]byte("null"))
	}))
	t.Cleanup(srv.Close)

	c, err := iikocloud.New(iikocloud.Config{
		APIKey: "k", AppID: "a", ClientSecret: "s", BaseURL: srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	return c, rec
}

// liveOperations is the generated table: the spec's own account of what exists,
// which is what the hand-written services are measured against.
func liveOperations() map[string]gen.Operation {
	out := make(map[string]gen.Operation, len(gen.Operations))
	for _, op := range gen.Operations {
		if op.Path == tokenPath {
			continue
		}
		out[op.Path] = op
	}
	return out
}

func services(c *iikocloud.Client) map[string]reflect.Value {
	out := map[string]reflect.Value{}
	v := reflect.ValueOf(c).Elem()
	for i := range v.NumField() {
		f := v.Type().Field(i)
		if f.IsExported() && v.Field(i).Kind() == reflect.Pointer {
			out[f.Name] = v.Field(i)
		}
	}
	return out
}

func TestEveryLiveOperationIsReachable(t *testing.T) {
	c, rec := dialFake(t)
	ctx := context.Background()

	covered := map[string]string{} // path -> Domain.Method
	for domain, svc := range services(c) {
		for i := range svc.NumMethod() {
			name := svc.Type().Method(i).Name
			args := []reflect.Value{reflect.ValueOf(ctx)}
			mt := svc.Method(i).Type()
			for a := 1; a < mt.NumIn(); a++ {
				args = append(args, reflect.New(mt.In(a)).Elem())
			}
			svc.Method(i).Call(args)

			path := rec.last()
			qualified := domain + "." + name
			if path == "" {
				t.Errorf("%s made no request", qualified)
				continue
			}
			if prev, dup := covered[path]; dup {
				t.Errorf("%s and %s both call %s", prev, qualified, path)
			}
			covered[path] = qualified
		}
	}

	var missing []string
	live := liveOperations()
	for path := range live {
		if _, ok := covered[path]; !ok {
			missing = append(missing, path)
		}
	}
	for path, by := range covered {
		if _, ok := live[path]; !ok {
			t.Errorf("%s calls %s, which is not a live operation", by, path)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d live operations have no method:\n%s", len(missing), strings.Join(missing, "\n"))
	}
}

func TestEveryMethodUsesTheSpecsRequestAndResponseTypes(t *testing.T) {
	c, rec := dialFake(t)
	ctx := context.Background()
	live := liveOperations()

	for domain, svc := range services(c) {
		for i := range svc.NumMethod() {
			name := domain + "." + svc.Type().Method(i).Name
			mt := svc.Method(i).Type()
			args := []reflect.Value{reflect.ValueOf(ctx)}
			for a := 1; a < mt.NumIn(); a++ {
				args = append(args, reflect.New(mt.In(a)).Elem())
			}
			svc.Method(i).Call(args)
			op, ok := live[rec.last()]
			if !ok {
				continue // reported by TestEveryLiveOperationIsReachable
			}

			wantReq := op.Request
			gotReq := ""
			if mt.NumIn() > 1 {
				gotReq = mt.In(1).Name()
			}
			if gotReq != wantReq {
				t.Errorf("%s takes %q, spec says %q", name, gotReq, wantReq)
			}
			if got := responseShape(mt); got != op.Response {
				t.Errorf("%s returns %q, spec says %q", name, got, op.Response)
			}
		}
	}
}

// responseShape renders a method's result the way the generated table spells it.
func responseShape(mt reflect.Type) string {
	if mt.NumOut() == 1 {
		return ""
	}
	rt := mt.Out(0)
	if rt.Kind() == reflect.Slice {
		return "[]" + rt.Elem().Name()
	}
	return rt.Elem().Name()
}

// TestRoundTripsEachMethodShape exercises the four shapes the spec produces with
// real bodies; the conformance tests above only check paths and types.
func TestRoundTripsEachMethodShape(t *testing.T) {
	var sent []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == tokenPath {
			json.NewEncoder(w).Encode(map[string]string{"token": "T"})
			return
		}
		sent, _ = io.ReadAll(r.Body)
		switch r.URL.Path {
		case "/api/1/organizations":
			// OrganizationInfo is a discriminated base, so the field is raw.
			org, _ := json.Marshal(gen.SimpleOrganizationInfo{
				OrganizationInfo: gen.OrganizationInfo{ID: "org-1", Name: "Ромашка", ResponseType: "Simple"},
			})
			json.NewEncoder(w).Encode(gen.GetOrganizationsResponse{
				CorrelationID: "c1", Organizations: []json.RawMessage{org},
			})
		case "/api/1/tips_types":
			json.NewEncoder(w).Encode(gen.TipsTypesResponse{CorrelationID: "c2"})
		case "/api/finance/v1/account-type/list":
			json.NewEncoder(w).Encode([]gen.AccountType{{Title: "Касса"}, {Title: "Банк"}})
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(srv.Close)

	c, err := iikocloud.New(iikocloud.Config{APIKey: "k", AppID: "a", ClientSecret: "s", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	orgs, err := c.Organizations.GetOrganizations(ctx, gen.GetOrganizationsRequest{OrganizationIDs: []string{"org-1"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(sent), `"organizationIds":["org-1"]`) {
		t.Errorf("request body was %s", sent)
	}
	var org gen.SimpleOrganizationInfo
	if len(orgs.Organizations) != 1 {
		t.Fatalf("got %+v", orgs.Organizations)
	}
	if err := json.Unmarshal(orgs.Organizations[0], &org); err != nil || org.Name != "Ромашка" {
		t.Errorf("got %+v (%v)", org, err)
	}

	// The two operations with no request body must still send valid JSON.
	if _, err := c.Organizations.GetTipsTypes(ctx); err != nil {
		t.Fatal(err)
	}
	if string(sent) != "{}" {
		t.Errorf("bodiless request sent %q", sent)
	}

	types, err := c.Finance.ListAccountType(ctx, gen.AccountTypeListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(types) != 2 || types[1].Title != "Банк" {
		t.Errorf("got %+v", types)
	}

	// A 200 with no body at all is a success, not a decode failure.
	if err := c.Menu.CreateNomenclatureProduct(ctx, gen.NomenclatureProductCreateRequest{}); err != nil {
		t.Fatal(err)
	}
}
