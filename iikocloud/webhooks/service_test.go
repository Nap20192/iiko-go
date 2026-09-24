package webhooks_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
	"github.com/Nap20192/iiko-go/iikocloud/rest"
	"github.com/Nap20192/iiko-go/iikocloud/webhooks"
)

// dial serves one canned operation. Path and type conformance for the whole
// surface is asserted once in pkg/iikocloud; this checks the wire round trip.
func dial(t *testing.T, path, response string, gotBody *string) *webhooks.Service {
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
	return webhooks.New(c)
}

func TestGetSettings(t *testing.T) {
	t.Parallel()

	var body string
	s := dial(t, "/api/1/webhooks/settings", `{"correlationId":"c-5","apiLoginName":"login","webHooksUri":"https://example.test/hook"}`, &body)

	got, err := s.GetSettings(context.Background(), gen.GetWebHookSettingsRequest{OrganizationID: "org-1"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, `"organizationId":"org-1"`) {
		t.Errorf("sent %s", body)
	}
	if got.WebHooksURI != "https://example.test/hook" {
		t.Errorf("got %+v", got)
	}
}
