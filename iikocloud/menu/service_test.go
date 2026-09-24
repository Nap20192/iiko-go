package menu_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
	"github.com/Nap20192/iiko-go/iikocloud/menu"
	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// dial serves one canned operation. Path and type conformance for the whole
// surface is asserted once in pkg/iikocloud; this checks the wire round trip.
func dial(t *testing.T, path, response string, gotBody *string) *menu.Service {
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
	return menu.New(c)
}

func TestGetStopLists(t *testing.T) {
	t.Parallel()

	var body string
	s := dial(t, "/api/1/stop_lists", `{"correlationId":"c-2","terminalGroupStopLists":[]}`, &body)

	got, err := s.GetStopLists(context.Background(), gen.StopListsRequest{OrganizationIDs: []string{"org-1"}, ReturnSize: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, `"returnSize":true`) {
		t.Errorf("sent %s", body)
	}
	if got.CorrelationID != "c-2" {
		t.Errorf("got %+v", got)
	}
}
