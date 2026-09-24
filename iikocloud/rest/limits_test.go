package rest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

func TestDocumentedLimitsMatchTheDocs(t *testing.T) {
	t.Parallel()

	// research/iiko-docs/ogranichenie-i-limity.txt, "на компанию". The Cloud API's
	// own per-group ceilings are stated to be undocumented, so only these appear.
	want := map[string]rest.Limit{
		"/api/1/loyalty/iiko/customer/info":                     {Every: 200 * time.Millisecond, Concurrent: 5},
		"/api/1/loyalty/iiko/customer/create_or_update":         {Every: 100 * time.Millisecond, Concurrent: 4},
		"/api/1/loyalty/iiko/customer/transactions/by_date":     {Every: 200 * time.Millisecond},
		"/api/1/loyalty/iiko/customer/transactions/by_revision": {Every: 200 * time.Millisecond},
		"/api/1/loyalty/iiko/coupons/by_series":                 {Every: 30 * time.Second},
	}
	for path, w := range want {
		got, ok := rest.DocumentedLimit(path)
		if !ok {
			t.Errorf("%s: no limit", path)
			continue
		}
		if got != w {
			t.Errorf("%s: got %+v, want %+v", path, got, w)
		}
	}
	if _, ok := rest.DocumentedLimit("/api/1/organizations"); ok {
		t.Error("unlisted paths must not carry an invented limit")
	}
}

func TestRateLimitedPathIsPaced(t *testing.T) {
	t.Parallel()

	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/access_token" {
			json.NewEncoder(w).Encode(map[string]string{"token": "T"})
			return
		}
		json.NewEncoder(w).Encode(echo{})
	})

	var slept []time.Duration
	now := time.Now()
	c := newClient(t, srv,
		rest.WithClock(func() time.Time { return now }),
		rest.WithSleep(func(_ context.Context, d time.Duration) error {
			slept = append(slept, d)
			return nil
		}))

	const path = "/api/1/loyalty/iiko/customer/info"
	for range 3 {
		if _, err := rest.Call[echo](context.Background(), c, path, nil); err != nil {
			t.Fatal(err)
		}
	}
	// The clock is frozen, so every call after the first must be spaced.
	want := []time.Duration{200 * time.Millisecond, 400 * time.Millisecond}
	if len(slept) != len(want) {
		t.Fatalf("slept %v, want %v", slept, want)
	}
	for i := range want {
		if slept[i] != want[i] {
			t.Fatalf("slept %v, want %v", slept, want)
		}
	}
}

func TestUnlimitedPathIsNotPaced(t *testing.T) {
	t.Parallel()

	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/access_token" {
			json.NewEncoder(w).Encode(map[string]string{"token": "T"})
			return
		}
		json.NewEncoder(w).Encode(echo{})
	})

	slept := 0
	now := time.Now()
	c := newClient(t, srv,
		rest.WithClock(func() time.Time { return now }),
		rest.WithSleep(func(context.Context, time.Duration) error { slept++; return nil }))

	for range 3 {
		rest.Call[echo](context.Background(), c, "/api/1/organizations", nil)
	}
	if slept != 0 {
		t.Fatalf("paced an unlimited path %d times", slept)
	}
}
