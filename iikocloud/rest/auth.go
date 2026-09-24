package rest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
)

// accessToken returns a live token, fetching one when the cached token is
// missing, inside the refresh window, or explicitly invalidated by a 401.
func (c *Client) accessToken(ctx context.Context, force bool) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !force && c.token != "" && c.now().Before(c.expires.Add(-refreshWindow)) {
		return c.token, nil
	}

	path, req := c.tokenRequest()
	body, err := encode(req)
	if err != nil {
		return "", err
	}
	status, respBody, header, err := c.send(ctx, path, body, "")
	if err != nil {
		return "", err
	}
	if status < 200 || status > 299 {
		return "", authError(parseError(status, path, respBody, header))
	}
	var out gen.GetAccessTokenV2Response
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", &AuthError{Err: &Error{
			Status: status, Path: path, Description: "token response is not JSON", Body: respBody,
		}}
	}
	if out.Token == "" {
		return "", &AuthError{Err: &Error{
			Status: status, Path: path, Description: "token response carried no token",
			CorrelationID: out.CorrelationID, Body: respBody,
		}}
	}

	c.token = out.Token
	c.expires = tokenExpiry(out.Token, c.now())
	return c.token, nil
}

func authError(err error) error {
	switch e := err.(type) {
	case *Error:
		return &AuthError{Err: e}
	case *RateLimitError:
		return &AuthError{Err: e.Err}
	default:
		return &AuthError{Err: &Error{Status: http.StatusInternalServerError, Path: tokenPath, Description: err.Error()}}
	}
}

// tokenExpiry reads the JWT exp claim the spec documents; anything unreadable
// falls back to the documented one-hour lifetime.
func tokenExpiry(token string, now time.Time) time.Time {
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		if raw, err := base64.RawURLEncoding.DecodeString(parts[1]); err == nil {
			var claims struct {
				Exp int64 `json:"exp"`
			}
			if json.Unmarshal(raw, &claims) == nil && claims.Exp > 0 {
				return time.Unix(claims.Exp, 0)
			}
		}
	}
	return now.Add(time.Hour)
}
