package rest_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// jwtExpiring builds a token whose exp claim the client is expected to read;
// only the payload segment is ever parsed, so the signature is a placeholder.
func jwtExpiring(at time.Time) string {
	payload, _ := json.Marshal(map[string]int64{"exp": at.Unix()})
	enc := base64.RawURLEncoding.EncodeToString
	return strings.Join([]string{enc([]byte(`{"alg":"HS256"}`)), enc(payload), "sig"}, ".")
}

func TestTokenRefreshesFiveMinutesBeforeExpiry(t *testing.T) {
	t.Parallel()

	var tokens atomic.Int32
	now := time.Now()
	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/access_token" {
			n := tokens.Add(1)
			json.NewEncoder(w).Encode(map[string]string{
				"token": jwtExpiring(now.Add(time.Hour)), "correlationId": "c",
			})
			_ = n
			return
		}
		json.NewEncoder(w).Encode(echo{})
	})

	clock := now
	c := newClient(t, srv, rest.WithClock(func() time.Time { return clock }))

	if _, err := rest.Call[echo](context.Background(), c, "/api/1/x", nil); err != nil {
		t.Fatal(err)
	}
	// Six minutes of headroom left: still valid.
	clock = now.Add(54 * time.Minute)
	if _, err := rest.Call[echo](context.Background(), c, "/api/1/x", nil); err != nil {
		t.Fatal(err)
	}
	if tokens.Load() != 1 {
		t.Fatalf("refreshed early: %d", tokens.Load())
	}
	// Four minutes left: inside the refresh window.
	clock = now.Add(56 * time.Minute)
	if _, err := rest.Call[echo](context.Background(), c, "/api/1/x", nil); err != nil {
		t.Fatal(err)
	}
	if tokens.Load() != 2 {
		t.Fatalf("did not refresh: %d", tokens.Load())
	}
}

func TestUnparsableTokenFallsBackToOneHour(t *testing.T) {
	t.Parallel()

	var tokens atomic.Int32
	now := time.Now()
	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/access_token" {
			tokens.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"token": "opaque", "correlationId": "c"})
			return
		}
		json.NewEncoder(w).Encode(echo{})
	})

	clock := now
	c := newClient(t, srv, rest.WithClock(func() time.Time { return clock }))
	rest.Call[echo](context.Background(), c, "/api/1/x", nil)
	clock = now.Add(56 * time.Minute)
	rest.Call[echo](context.Background(), c, "/api/1/x", nil)
	if tokens.Load() != 2 {
		t.Fatalf("documented TTL is one hour; refreshed %d times", tokens.Load())
	}
}

func TestOn401RefreshesAndRetriesExactlyOnce(t *testing.T) {
	t.Parallel()

	var tokens, calls atomic.Int32
	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/access_token" {
			n := tokens.Add(1)
			json.NewEncoder(w).Encode(map[string]string{
				"token": fmt.Sprintf("T%d", n), "correlationId": "c"})
			return
		}
		n := calls.Add(1)
		if n == 1 {
			if got := r.Header.Get("Authorization"); got != "Bearer T1" {
				t.Errorf("first call used %q", got)
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"errorDescription": "expired", "correlationId": "c1"})
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer T2" {
			t.Errorf("retry used %q, want the refreshed token", got)
		}
		json.NewEncoder(w).Encode(echo{Value: "ok"})
	})

	c := newClient(t, srv)
	got, err := rest.Call[echo](context.Background(), c, "/api/1/x", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Value != "ok" {
		t.Fatalf("got %q", got.Value)
	}
	if calls.Load() != 2 {
		t.Fatalf("made %d calls, want exactly one retry", calls.Load())
	}
}

func TestPersistent401StopsAfterOneRetry(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/access_token" {
			json.NewEncoder(w).Encode(map[string]string{"token": "T", "correlationId": "c"})
			return
		}
		calls.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"errorDescription": "nope", "correlationId": "c9"})
	})

	c := newClient(t, srv)
	if _, err := rest.Call[echo](context.Background(), c, "/api/1/x", nil); err == nil {
		t.Fatal("want an error")
	}
	if calls.Load() != 2 {
		t.Fatalf("made %d calls, want 2", calls.Load())
	}
}

func TestRefreshFailureIsTyped(t *testing.T) {
	t.Parallel()

	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorDescription": "bad clientSecret", "error": "InvalidCredentials", "correlationId": "c7"})
	})

	c := newClient(t, srv)
	_, err := rest.Call[echo](context.Background(), c, "/api/1/x", nil)
	var authErr *rest.AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("want *rest.AuthError, got %T: %v", err, err)
	}
	if authErr.Err.CorrelationID != "c7" || authErr.Err.Description != "bad clientSecret" {
		t.Fatalf("lost detail: %+v", authErr.Err)
	}
}

// A restaurant that only creates an API key in iikoWeb has an apiLogin and no
// developer-portal application, so the v1 token endpoint is the only one it can
// use. Config.APILogin selects it.
func TestAPILoginUsesTheV1TokenEndpoint(t *testing.T) {
	t.Parallel()

	var path, body string
	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "access_token") {
			path = r.URL.Path
			b, _ := io.ReadAll(r.Body)
			body = string(b)
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "T"})
			return
		}
		_ = json.NewEncoder(w).Encode(echo{})
	})
	c, err := rest.New(rest.Config{APILogin: "login-1", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rest.Call[echo](context.Background(), c, "/api/1/x", nil); err != nil {
		t.Fatal(err)
	}
	if path != "/api/1/access_token" {
		t.Errorf("token path = %q", path)
	}
	if !strings.Contains(body, `"apiLogin":"login-1"`) {
		t.Errorf("token body = %s", body)
	}
}

func TestConfigNeedsOneCredentialSet(t *testing.T) {
	t.Parallel()

	if _, err := rest.New(rest.Config{}); err == nil {
		t.Error("empty config must be refused")
	}
	if _, err := rest.New(rest.Config{APIKey: "k"}); err == nil {
		t.Error("a partial v2 credential set must be refused")
	}
	// apiLogin alone is enough; so is the full v2 triple.
	if _, err := rest.New(rest.Config{APILogin: "l"}); err != nil {
		t.Errorf("apiLogin alone: %v", err)
	}
	if _, err := rest.New(rest.Config{APIKey: "k", AppID: "a", ClientSecret: "s"}); err != nil {
		t.Errorf("v2 triple: %v", err)
	}
}
