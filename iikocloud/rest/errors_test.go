package rest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

func TestErrorFieldPrecedence(t *testing.T) {
	t.Parallel()

	// iikoCloud spells the same failure three ways depending on which service
	// answered; the most specific wording wins.
	cases := []struct {
		name string
		body string
		want string
	}{
		{"errorDescription first", `{"errorDescription":"desc","error":"code","message":"msg"}`, "desc"},
		{"error second", `{"error":"code","message":"msg"}`, "code"},
		{"message last", `{"message":"msg"}`, "msg"},
		{"nothing usable", `{}`, "400 Bad Request"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v2/access_token" {
					json.NewEncoder(w).Encode(map[string]string{"token": "T"})
					return
				}
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(tc.body))
			})
			_, err := rest.Call[echo](context.Background(), newClient(t, srv), "/api/1/x", nil)
			var e *rest.Error
			if !errors.As(err, &e) {
				t.Fatalf("want *rest.Error, got %T: %v", err, err)
			}
			if e.Description != tc.want {
				t.Fatalf("Description = %q, want %q", e.Description, tc.want)
			}
			if e.Status != http.StatusBadRequest {
				t.Fatalf("Status = %d", e.Status)
			}
		})
	}
}

func TestErrorCarriesCorrelationID(t *testing.T) {
	t.Parallel()

	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/access_token" {
			json.NewEncoder(w).Encode(map[string]string{"token": "T"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"errorDescription": "boom", "correlationId": "abc-123"})
	})

	_, err := rest.Call[echo](context.Background(), newClient(t, srv), "/api/1/x", nil)
	var e *rest.Error
	if !errors.As(err, &e) {
		t.Fatalf("got %T", err)
	}
	// Support asks for it by name; losing it makes a report unactionable.
	if e.CorrelationID != "abc-123" {
		t.Fatalf("CorrelationID = %q", e.CorrelationID)
	}
}

func TestTooManyRequestsIsTypedAndNotRetried(t *testing.T) {
	t.Parallel()

	calls := 0
	srv := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/access_token" {
			json.NewEncoder(w).Encode(map[string]string{"token": "T"})
			return
		}
		calls++
		w.Header().Set("Retry-After", "45")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{"errorDescription": "slot exhausted", "correlationId": "c"})
	})

	_, err := rest.Call[echo](context.Background(), newClient(t, srv), "/api/1/x", nil)
	var e *rest.RateLimitError
	if !errors.As(err, &e) {
		t.Fatalf("want *rest.RateLimitError, got %T: %v", err, err)
	}
	if e.RetryAfter != 45*time.Second {
		t.Fatalf("RetryAfter = %v", e.RetryAfter)
	}
	// A slot lasts up to an hour; hammering it is itself grounds for a block.
	if calls != 1 {
		t.Fatalf("made %d calls, want 1", calls)
	}
}

func TestErrorMessagesCarryWhatSupportAsksFor(t *testing.T) {
	t.Parallel()

	base := &rest.Error{Status: 500, Path: "/api/1/x", Description: "boom", CorrelationID: "c-1"}
	cases := map[string]error{
		"plain": base,
		"rate":  &rest.RateLimitError{Err: base, RetryAfter: 45 * time.Second},
		"auth":  &rest.AuthError{Err: base},
	}
	for name, err := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			msg := err.Error()
			for _, want := range []string{"/api/1/x", "500", "boom", "c-1"} {
				if !strings.Contains(msg, want) {
					t.Errorf("%q missing from %q", want, msg)
				}
			}
			var unwrapped *rest.Error
			if !errors.As(err, &unwrapped) || unwrapped != base {
				t.Errorf("does not unwrap to the underlying *rest.Error")
			}
		})
	}
}
