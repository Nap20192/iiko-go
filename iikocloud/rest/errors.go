package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Error is a failed iikoCloud call. CorrelationID is what iiko support asks for.
type Error struct {
	Status        int
	Path          string
	Code          string
	Description   string
	CorrelationID string
	Body          []byte
}

func (e *Error) Error() string {
	return fmt.Sprintf("iikocloud %s: %d %s (correlationId %s)", e.Path, e.Status, e.Description, e.CorrelationID)
}

// RateLimitError is HTTP 429. The slot it names lasts up to an hour, so the
// client surfaces it rather than retrying into a permanent API-login block.
type RateLimitError struct {
	Err        *Error
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("%s: rate limited, retry after %s", e.Err.Error(), e.RetryAfter)
}
func (e *RateLimitError) Unwrap() error { return e.Err }

// AuthError is a failure to obtain a token, as opposed to a rejected call.
type AuthError struct {
	Err *Error
}

func (e *AuthError) Error() string { return "iikocloud auth: " + e.Err.Error() }
func (e *AuthError) Unwrap() error { return e.Err }

// parseError reads the three spellings iikoCloud uses for the same failure.
func parseError(status int, path string, body []byte, header http.Header) error {
	var payload struct {
		ErrorDescription string `json:"errorDescription"`
		Error            string `json:"error"`
		Message          string `json:"message"`
		CorrelationID    string `json:"correlationId"`
	}
	_ = json.Unmarshal(body, &payload)

	desc := payload.ErrorDescription
	if desc == "" {
		desc = payload.Error
	}
	if desc == "" {
		desc = payload.Message
	}
	if desc == "" {
		desc = fmt.Sprintf("%d %s", status, http.StatusText(status))
	}

	e := &Error{
		Status: status, Path: path, Code: payload.Error,
		Description: desc, CorrelationID: payload.CorrelationID, Body: body,
	}
	if status == http.StatusTooManyRequests {
		return &RateLimitError{Err: e, RetryAfter: retryAfter(header)}
	}
	return e
}

func retryAfter(h http.Header) time.Duration {
	secs, err := strconv.Atoi(h.Get("Retry-After"))
	if err != nil || secs < 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}
