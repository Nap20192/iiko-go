package customers_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Nap20192/iiko-go/iikocloud/customers"
	"github.com/Nap20192/iiko-go/iikocloud/gen"
	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// dial serves one canned operation. Path and type conformance for the whole
// surface is asserted once in pkg/iikocloud; this checks the wire round trip.
func dial(t *testing.T, path, response string, gotBody *string) *customers.Service {
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
	return customers.New(c)
}

func TestGetCustomerInfo(t *testing.T) {
	t.Parallel()

	var body string
	s := dial(t, "/api/1/loyalty/iiko/customer/info", `{"id":"guest-1","name":"Иван","walletBalances":[]}`, &body)

	got, err := s.GetCustomerInfo(context.Background(), gen.GetCustomerInfoRequest{OrganizationID: "org-1", Type: "phone"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, `"type":"phone"`) {
		t.Errorf("sent %s", body)
	}
	if got.Name != "Иван" {
		t.Errorf("got %+v", got)
	}
}
