package rest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// newFake stands in for iikoCloud. There is no cloud demo stand, so this is the
// only boundary the client is ever exercised against.
func newFake(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func newClient(t *testing.T, srv *httptest.Server, opts ...rest.Option) *rest.Client {
	t.Helper()
	c, err := rest.New(rest.Config{
		APIKey: "key", AppID: "app", ClientSecret: "secret", BaseURL: srv.URL,
	}, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

type echo struct {
	Value string `json:"value"`
}

func TestPostAuthenticatesOnceAndSendsBearer(t *testing.T) {
	t.Parallel()

	var tokens, calls atomic.Int32
	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/access_token" {
			tokens.Add(1)
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["apiKey"] != "key" || body["appId"] != "app" || body["clientSecret"] != "secret" {
				t.Errorf("credentials not sent: %v", body)
			}
			if got := r.Header.Get("Authorization"); got != "" {
				t.Errorf("token request must be unauthenticated, got %q", got)
			}
			json.NewEncoder(w).Encode(map[string]string{"token": "T1", "correlationId": "c"})
			return
		}
		calls.Add(1)
		if got := r.Header.Get("Authorization"); got != "Bearer T1" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		var in echo
		json.NewDecoder(r.Body).Decode(&in)
		json.NewEncoder(w).Encode(echo{Value: in.Value + "!"})
	})

	c := newClient(t, srv)
	for range 2 {
		got, err := rest.Call[echo](context.Background(), c, "/api/1/x", echo{Value: "hi"})
		if err != nil {
			t.Fatal(err)
		}
		if got.Value != "hi!" {
			t.Fatalf("got %q", got.Value)
		}
	}
	if tokens.Load() != 1 {
		t.Errorf("authenticated %d times, want 1", tokens.Load())
	}
	if calls.Load() != 2 {
		t.Errorf("made %d calls, want 2", calls.Load())
	}
}

func TestCallListDecodesABareArray(t *testing.T) {
	t.Parallel()

	// Sixteen operations answer with a bare JSON array rather than a wrapper.
	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/access_token" {
			json.NewEncoder(w).Encode(map[string]string{"token": "T"})
			return
		}
		w.Write([]byte(`[{"value":"a"},{"value":"b"}]`))
	})

	got, err := rest.CallList[echo](context.Background(), newClient(t, srv), "/api/1/list", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Value != "a" || got[1].Value != "b" {
		t.Fatalf("got %v", got)
	}
}

func TestPacingHonoursContextCancellation(t *testing.T) {
	t.Parallel()

	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"token": "T"})
	})
	// The real sleeper is in play here: a 30-second coupon slot must not pin the
	// caller once its context is done.
	c := newClient(t, srv, rest.WithHTTPClient(&http.Client{}))
	ctx, cancel := context.WithCancel(context.Background())

	const path = "/api/1/loyalty/iiko/coupons/by_series"
	if _, err := rest.Call[echo](ctx, c, path, nil); err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err := rest.Call[echo](ctx, c, path, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}
