// Package resttest builds a rest.Client wired to a stub iiko, so a domain test
// only has to describe the endpoint it cares about.
package resttest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// NewFake serves h for every request except auth, which always succeeds.
func NewFake(t *testing.T, h http.HandlerFunc) *rest.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, rest.EndpointAuth) {
			_, _ = w.Write([]byte("tok"))
			return
		}
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	return rest.New(rest.Config{BaseURL: srv.URL + "/resto", Login: "u", Password: "p"})
}
