// Package iiko is a client for the iikoServer back-office API (/resto/api).
// It has no MCP dependencies: the transport layer is a thin shell over this.
package rest

import (
	"context"
	"crypto/sha1" //nolint:gosec // iiko mandates SHA1 for the auth digest; not our choice
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Config is read from the environment; nothing here ever reaches a tool argument.
type Config struct {
	BaseURL    string // e.g. https://host:8080/resto  or https://<crm-id>.iiko.it/resto
	Login      string
	Password   string // plaintext; hashed on the wire
	SkipTLS    bool   // hosted iiko.it stands commonly ship self-signed certs
	Timeout    time.Duration
	AllowWrite bool
}

// Client holds exactly one API session for its whole lifetime.
//
// Two invariants come straight from the iiko docs and are not negotiable:
//
//  1. Each successful /auth occupies one licence slot. A customer with a
//     single-seat API licence can hold exactly ONE token; a second auth fails
//     with 403 and, worse, a leaked token keeps a real cashier locked out of
//     iikoOffice. So: one token per process, re-auth only on 401, and always
//     logout on shutdown.
//  2. "Запросы должны выполняться последовательно" — the server requires
//     strictly sequential requests. MCP runtimes call tools concurrently, so
//     every request serializes through mu.
type Client struct {
	cfg  Config
	http *http.Client

	mu    sync.Mutex // serializes ALL requests; also guards token
	token string
}

func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 120 * time.Second // OLAP over a wide range is genuinely slow
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	tr := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.SkipTLS {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // opt-in, self-signed iiko.it stands
	}
	// Since 4.3 the server sets the session as a `key` cookie itself. Carrying
	// the jar means the token travels both ways it can, which matters on builds
	// that stop honouring the query parameter.
	jar, err := cookiejar.New(nil)
	if err != nil {
		// cookiejar.New(nil) cannot fail; degrade to query-param-only rather
		// than panicking in a library constructor.
		jar = nil
	}
	return &Client{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout, Transport: tr, Jar: jar}}
}

func (c *Client) AllowWrite() bool { return c.cfg.AllowWrite }

// Error carries the iiko failure plus a recovery hint aimed at an LLM caller.
// Tool layers surface this as an isError result, never as a protocol error, so
// the model can see it and self-correct.
type Error struct {
	Status int
	Body   string
	Hint   string
}

func (e *Error) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("iiko %d: %s\n%s", e.Status, e.Body, e.Hint)
	}
	return fmt.Sprintf("iiko %d: %s", e.Status, e.Body)
}

// hintFor translates an iiko status into a named recovery action.
// 409 bodies are written for end users and are the only localized ones, so we
// pass them through verbatim and add nothing.
func hintFor(status int, body string) string {
	switch status {
	case http.StatusBadRequest:
		return "Fix the request parameters. Date formats differ per endpoint family — v1 /reports/* want DD.MM.YYYY, v2 wants yyyy-MM-dd, OLAP filters want yyyy-MM-ddTHH:mm:ss.SSS."
	case http.StatusUnauthorized:
		return "Session expired. Retry the call — the client re-authenticates automatically."
	case http.StatusForbidden:
		if strings.Contains(body, "no connections available") {
			return "No free licence slots. Close an iikoOffice session or raise the API licence. Check free slots with server_info."
		}
		if strings.Contains(body, "blocked within current license") {
			return "That licence module is not enabled on this server."
		}
		return "Authenticated but not permitted. The API user is missing a right (e.g. B_EN for nomenclature, B_VTJ + module 2200 for events)."
	case http.StatusNotFound:
		return "Object or path not found. Verify ids with list_entities."
	case http.StatusConflict:
		return "" // body is already an end-user message; adding to it only adds noise
	case http.StatusInternalServerError:
		return "iiko internal error. Narrow the date range and retry; if it persists the server log has more than the API exposes."
	}
	return ""
}

// do runs one request with the shared token, re-authenticating once on 401.
// The mutex is held for the whole call — see Client's invariant 2.
func (c *Client) Do(ctx context.Context, method, path string, q url.Values, body io.Reader, contentType string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var seekable []byte
	if body != nil {
		var err error
		if seekable, err = io.ReadAll(body); err != nil {
			return nil, fmt.Errorf("read request body: %w", err)
		}
	}

	if c.token == "" {
		if err := c.authLocked(ctx); err != nil {
			return nil, err
		}
	}

	data, status, err := c.attempt(ctx, method, path, q, seekable, contentType)
	if err != nil {
		return nil, err
	}
	if status == http.StatusUnauthorized {
		// Token died (1h idle, or the server restarted). Re-auth once, retry once.
		c.token = ""
		if err := c.authLocked(ctx); err != nil {
			return nil, err
		}
		data, status, err = c.attempt(ctx, method, path, q, seekable, contentType)
		if err != nil {
			return nil, err
		}
	}
	if status < 200 || status > 299 {
		return nil, &Error{Status: status, Body: strings.TrimSpace(string(data)), Hint: hintFor(status, string(data))}
	}
	// A text/html body means a network or config problem (wrong URL, proxy, bare
	// Tomcat) rather than an API response — the docs call this out explicitly.
	if strings.HasPrefix(strings.TrimSpace(string(data)), "<!DOCTYPE html") {
		return nil, &Error{Status: status, Body: "server returned HTML, not an API response",
			Hint: "Check IIKO_BASE_URL — it must end in /resto and point at the RMS app server, not a proxy or a bare Tomcat."}
	}
	return data, nil
}

func (c *Client) attempt(ctx context.Context, method, path string, q url.Values, body []byte, contentType string) (data []byte, status int, err error) {
	if q == nil {
		q = url.Values{}
	}
	q.Set("key", c.token)
	u := c.cfg.BaseURL + path + "?" + q.Encode()

	var rdr io.Reader
	if body != nil {
		rdr = strings.NewReader(string(body))
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rdr)
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("call %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}
	return data, resp.StatusCode, nil
}

// authLocked must be called with mu held.
func (c *Client) authLocked(ctx context.Context) error {
	sum := sha1.Sum([]byte(c.cfg.Password)) //nolint:gosec // iiko mandates SHA1 here
	q := url.Values{"login": {c.cfg.Login}, "pass": {hex.EncodeToString(sum[:])}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.BaseURL+EndpointAuth+"?"+q.Encode(), http.NoBody)
	if err != nil {
		return fmt.Errorf("build auth request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read auth response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return &Error{Status: resp.StatusCode, Body: strings.TrimSpace(string(data)),
			Hint: hintFor(resp.StatusCode, string(data))}
	}
	tok := strings.TrimSpace(string(data))
	if tok == "" || strings.ContainsAny(tok, " <\n") {
		return &Error{Status: resp.StatusCode, Body: "auth returned no usable token",
			Hint: "Check IIKO_LOGIN / IIKO_PASSWORD and that the user has API access."}
	}
	c.token = tok
	return nil
}

// Close releases the licence slot. Not calling this leaves a seat occupied
// until the ~1h idle timeout, which can lock a real cashier out of iikoOffice.
func (c *Client) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token == "" {
		return nil
	}
	tok := c.token
	c.token = ""
	if c.http.Jar != nil {
		// Drop the server-set cookie too; otherwise a dead session id keeps
		// riding along on any later request.
		if u, err := url.Parse(c.cfg.BaseURL); err == nil {
			c.http.Jar.SetCookies(u, nil)
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.cfg.BaseURL+EndpointLogout+"?"+url.Values{"key": {tok}}.Encode(), http.NoBody)
	if err != nil {
		return fmt.Errorf("build logout request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// GetXML and GetV2/PostV2 are the two response worlds of resto/api. The split is
// by endpoint, not cleanly by path prefix: /corporation/settings and
// /entities/accounts/list are JSON despite v1-looking paths.
//
// There is deliberately no plain GetJSON/PostJSON. Every JSON endpoint this
// server touches is under /api/v2/, and a raw json.Unmarshal there would decode
// an HTTP 200 result=ERROR body into an empty result and report success. Routing
// all JSON through DecodeV2 makes that mistake unavailable rather than merely
// discouraged.
func (c *Client) GetXML(ctx context.Context, path string, q url.Values, out any) error {
	data, err := c.Do(ctx, http.MethodGet, path, q, nil, "")
	if err != nil {
		return err
	}
	if err := xml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode XML from %s: %w", path, err)
	}
	return nil
}

// PostV2 posts to a v2 JSON endpoint and normalizes the envelope asymmetry.
// Needed because iiko reports business failures as result=ERROR with HTTP 200,
// so a plain json.Unmarshal of an OLAP response decodes an error body into an
// empty report and reports success. See DecodeV2.
func (c *Client) PostV2(ctx context.Context, path string, q url.Values, in, out any) (*int64, error) {
	raw, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	data, err := c.Do(ctx, http.MethodPost, path, q, strings.NewReader(string(raw)),
		"application/json; charset=utf-8")
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil // a caller that wants no body still wants the error check above
	}
	rev, err := DecodeV2(data, out)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return rev, nil
}

// PostXML posts an XML body and decodes an XML response.
func (c *Client) PostXML(ctx context.Context, path string, q url.Values, in, out any) error {
	body, err := xml.Marshal(in)
	if err != nil {
		return fmt.Errorf("encode XML request: %w", err)
	}
	data, err := c.Do(ctx, http.MethodPost, path, q, strings.NewReader(xml.Header+string(body)), "application/xml")
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	// G709: Go's encoding/xml resolves no external entities, so there is no XXE
	// here; upstream free text is neutralised later by render.Sanitize.
	if err := xml.Unmarshal(data, out); err != nil { //nolint:gosec // see above
		return fmt.Errorf("decode XML from %s: %w", path, err)
	}
	return nil
}

// PutXML replaces an entity wholesale. iiko's REST convention is that PUT is a
// full replace: optional fields left out are cleared, not kept.
func (c *Client) PutXML(ctx context.Context, path string, q url.Values, in, out any) error {
	body, err := xml.Marshal(in)
	if err != nil {
		return fmt.Errorf("encode XML request: %w", err)
	}
	data, err := c.Do(ctx, http.MethodPut, path, q, strings.NewReader(xml.Header+string(body)), "application/xml")
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	// G709: see PostXML — encoding/xml resolves no external entities.
	if err := xml.Unmarshal(data, out); err != nil { //nolint:gosec // see above
		return fmt.Errorf("decode XML from %s: %w", path, err)
	}
	return nil
}

// GetV2 fetches a v2 JSON endpoint and normalizes the envelope asymmetry.
// Use it for anything under /api/v2/ that is a GET.
func (c *Client) GetV2(ctx context.Context, path string, q url.Values, out any) (*int64, error) {
	data, err := c.Do(ctx, http.MethodGet, path, q, nil, "")
	if err != nil {
		return nil, err
	}
	rev, err := DecodeV2(data, out)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return rev, nil
}

// BoolStr renders a bool for a query param; iiko defaults flip per endpoint, so
// callers send it explicitly rather than omitting it.
func BoolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// IsNotFound reports whether err is an iiko 404, used to fall back between the
// two documented spellings of a path.
func IsNotFound(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Status == http.StatusNotFound
}

// LicenceInfo reports free slots for a licence module. Sessions are licence
// seats, not rate-limit slots, so this is a session concern.
func (c *Client) LicenceInfo(ctx context.Context, moduleID string) (string, error) {
	data, err := c.Do(ctx, http.MethodGet, EndpointLicence, url.Values{"moduleId": {moduleID}}, nil, "")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
