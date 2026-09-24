package inventory_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
	"github.com/Nap20192/iiko-go/iikocloud/inventory"
	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// dial serves one canned operation. Path and type conformance for the whole
// surface is asserted once in pkg/iikocloud; this checks the wire round trip.
func dial(t *testing.T, path, response string, gotBody *string) *inventory.Service {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/access_token" {
			json.NewEncoder(w).Encode(map[string]string{"token": "T"})
			return
		}
		if r.URL.Path != path {
			t.Errorf("called %s, want %s", r.URL.Path, path)
		}
		b, _ := io.ReadAll(r.Body)
		*gotBody = string(b)
		w.Write([]byte(response))
	}))
	t.Cleanup(srv.Close)

	c, err := rest.New(rest.Config{APIKey: "k", AppID: "a", ClientSecret: "s", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	return inventory.New(c)
}

func TestListStockBalance(t *testing.T) {
	t.Parallel()

	var body string
	s := dial(t, "/api/inventory/v1/stock_balance/list", `{"balanceAt":"2026-09-23 00:00:00.000","items":[{"productId":"p-1","quantity":3.5}],"limit":100}`, &body)

	got, err := s.ListStockBalance(context.Background(), gen.StockBalanceListRequest{BalanceAt: "2026-09-23 00:00:00.000", OrganizationID: "org-1"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, `"balanceAt":"2026-09-23 00:00:00.000"`) {
		t.Errorf("sent %s", body)
	}
	if len(got.Items) != 1 || got.Items[0].Quantity != 3.5 {
		t.Errorf("got %+v", got.Items)
	}
}
