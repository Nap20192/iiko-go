// Package rest is the iikoCloud transport: bearer tokens, the documented rate
// limits, and iikoCloud's three spellings of an error. It knows no domain.
package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
)

// Config is the credential set from the iiko developer portal plus the API key
// the restaurant generates in iikoWeb.
type Config struct {
	// APILogin is the API key created in iikoWeb under "Настройка Cloud API".
	// A restaurant with only this key authenticates through the v1 token
	// endpoint; APIKey/AppID/ClientSecret then stay empty. When both are set,
	// APILogin wins.
	APILogin     string
	APIKey       string
	AppID        string
	ClientSecret string

	// BaseURL defaults to the spec's own servers entry.
	BaseURL string
	// Timeout, when set, becomes the API's own per-request Timeout header.
	Timeout time.Duration
}

// Option overrides a default that only a test or an unusual deployment needs.
type Option func(*Client)

// WithHTTPClient replaces the transport, for proxies and instrumentation.
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }

// WithClock freezes time so token expiry and pacing are testable.
func WithClock(now func() time.Time) Option { return func(c *Client) { c.now = now } }

// WithSleep replaces the pacing wait, so tests observe it instead of serving it.
func WithSleep(f func(context.Context, time.Duration) error) Option {
	return func(c *Client) { c.sleep = f }
}

// Client is one authenticated iikoCloud session.
type Client struct {
	cfg   Config
	http  *http.Client
	now   func() time.Time
	sleep func(context.Context, time.Duration) error
	lim   *limiter

	mu      sync.Mutex
	token   string
	expires time.Time
}

// refreshWindow is the headroom the client keeps: the documented TTL is an hour
// and there is no refresh-token flow, so a token is replaced before it expires.
const refreshWindow = 5 * time.Minute

//nolint:gosec // URL paths, not credentials
const (
	tokenPath   = "/api/v2/access_token"
	tokenPathV1 = "/api/1/access_token"
)

// tokenRequest is the endpoint and body for the credentials in use.
func (c *Client) tokenRequest() (path string, body any) {
	if c.cfg.APILogin != "" {
		return tokenPathV1, gen.GetAccessTokenRequest{APILogin: c.cfg.APILogin}
	}
	return tokenPath, gen.GetAccessTokenV2Request{
		APIKey: c.cfg.APIKey, AppID: c.cfg.AppID, ClientSecret: c.cfg.ClientSecret,
	}
}

// New builds a client. It performs no network call; the token is fetched lazily.
func New(cfg Config, opts ...Option) (*Client, error) {
	if cfg.APILogin == "" && (cfg.APIKey == "" || cfg.AppID == "" || cfg.ClientSecret == "") {
		return nil, fmt.Errorf("iikocloud: either apiLogin, or apiKey, appId and clientSecret together, are required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = gen.DefaultBaseURL
	}
	c := &Client{
		cfg:   cfg,
		http:  &http.Client{Timeout: 2 * time.Minute},
		now:   time.Now,
		sleep: sleepCtx,
		lim:   newLimiter(),
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Call posts a request and decodes the response. Every iikoCloud operation is a
// POST of JSON returning JSON, so this is the only shape a service needs.
func Call[Resp any](ctx context.Context, c *Client, path string, body any) (*Resp, error) {
	var out Resp
	if err := c.Post(ctx, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CallList is Call for the operations whose 200 body is a bare JSON array.
func CallList[Item any](ctx context.Context, c *Client, path string, body any) ([]Item, error) {
	out, err := Call[[]Item](ctx, c, path, body)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// Post is Call without the generic, for operations with no response body.
func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	if wait := c.lim.reserve(path, c.now()); wait > 0 {
		if err := c.sleep(ctx, wait); err != nil {
			return err
		}
	}
	release, err := c.lim.acquire(ctx, path)
	if err != nil {
		return err
	}
	defer release()

	payload, err := encode(body)
	if err != nil {
		return err
	}
	token, err := c.accessToken(ctx, false)
	if err != nil {
		return err
	}

	status, respBody, header, err := c.send(ctx, path, payload, token)
	if err != nil {
		return err
	}
	// Exactly one retry, and only after a forced refresh: a second 401 is a
	// credential problem, not an expiry, and retrying it earns a block.
	if status == http.StatusUnauthorized {
		if token, err = c.accessToken(ctx, true); err != nil {
			return err
		}
		if status, respBody, header, err = c.send(ctx, path, payload, token); err != nil {
			return err
		}
	}
	if status < 200 || status > 299 {
		return parseError(status, path, respBody, header)
	}
	if out == nil || len(bytes.TrimSpace(respBody)) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("iikocloud %s: decode response: %w", path, err)
	}
	return nil
}

func (c *Client) send(ctx context.Context, path string, body []byte, token string) (status int, respBody []byte, header http.Header, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if c.cfg.Timeout > 0 {
		req.Header.Set("Timeout", strconv.Itoa(int(c.cfg.Timeout.Seconds())))
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("iikocloud %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("iikocloud %s: read response: %w", path, err)
	}
	return resp.StatusCode, respBody, resp.Header, nil
}

func encode(body any) ([]byte, error) {
	if body == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(body)
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
