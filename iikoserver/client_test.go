package iikoserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

func TestNewWiresEveryDomainToOneTransport(t *testing.T) {
	t.Parallel()
	c := New(Config{BaseURL: "http://127.0.0.1:1/resto"})
	domains := map[string]any{
		"Corporation": c.Corporation, "Nomenclature": c.Nomenclature,
		"Recipes": c.Recipes, "Reports": c.Reports,
		"Cashshifts": c.Cashshifts, "Documents": c.Documents,
		"Staff": c.Staff, "Suppliers": c.Suppliers,
	}
	for name, d := range domains {
		if d == nil {
			t.Errorf("%s is nil; New must wire every domain", name)
		}
	}
}

// The facade's Close is the one callers reach for, and a leaked token holds a
// licence seat until the idle timeout.
func TestCloseReleasesTheLicenceSeatOnce(t *testing.T) {
	t.Parallel()
	var auths, logouts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, rest.EndpointAuth):
			auths.Add(1)
			_, _ = w.Write([]byte("tok"))
		case strings.HasSuffix(r.URL.Path, rest.EndpointLogout):
			logouts.Add(1)
			_, _ = w.Write([]byte("ok"))
		default:
			_, _ = w.Write([]byte(`<corporateItemDtoes></corporateItemDtoes>`))
		}
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL + "/resto", Login: "u", Password: "p"})
	ctx := context.Background()
	if _, err := c.Corporation.ListEntities(ctx, "departments", false); err != nil {
		t.Fatalf("call through the facade: %v", err)
	}
	if got := auths.Load(); got != 1 {
		t.Fatalf("one session means one licence seat, saw %d auths", got)
	}
	if err := c.Close(ctx); err != nil {
		t.Fatalf("close: %v", err)
	}
	if logouts.Load() != 1 {
		t.Fatal("Close must log out: a leaked token can lock a cashier out of iikoOffice")
	}
	if err := c.Close(ctx); err != nil || logouts.Load() != 1 {
		t.Fatalf("second Close must be a no-op, saw %d logouts (%v)", logouts.Load(), err)
	}
}

func TestAllowWriteReflectsConfig(t *testing.T) {
	t.Parallel()
	if New(Config{}).AllowWrite() {
		t.Error("writes must be off unless configured")
	}
	if !New(Config{AllowWrite: true}).AllowWrite() {
		t.Error("AllowWrite must reflect config through the facade")
	}
}
