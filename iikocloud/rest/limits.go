package rest

import (
	"context"
	"sync"
	"time"
)

// Limit is a documented ceiling for one method group, per company.
type Limit struct {
	Every      time.Duration // minimum spacing between requests
	Concurrent int           // simultaneous requests, 0 for unstated
}

// documented holds the only per-method numbers iiko publishes
// (research/iiko-docs/ogranichenie-i-limity.txt); the Cloud API's own per-group
// ceilings are set "экспертно" and never printed, so nothing else is paced.
var documented = map[string]Limit{
	"/api/1/loyalty/iiko/customer/info":                     {Every: time.Second / 5, Concurrent: 5},
	"/api/1/loyalty/iiko/customer/create_or_update":         {Every: time.Second / 10, Concurrent: 4},
	"/api/1/loyalty/iiko/customer/transactions/by_date":     {Every: time.Second / 5},
	"/api/1/loyalty/iiko/customer/transactions/by_revision": {Every: time.Second / 5},
	"/api/1/loyalty/iiko/coupons/by_series":                 {Every: 30 * time.Second},
}

// globalConcurrent is the loyalty server's overall ceiling on in-flight requests.
const globalConcurrent = 100

// DocumentedLimit reports the published limit for a path, if iiko published one.
func DocumentedLimit(path string) (Limit, bool) {
	l, ok := documented[path]
	return l, ok
}

type limiter struct {
	mu     sync.Mutex
	next   map[string]time.Time
	slots  map[string]chan struct{}
	global chan struct{}
}

func newLimiter() *limiter {
	return &limiter{
		next:   map[string]time.Time{},
		slots:  map[string]chan struct{}{},
		global: make(chan struct{}, globalConcurrent),
	}
}

// reserve claims the next slot for a path and returns how long to wait for it.
func (l *limiter) reserve(path string, now time.Time) time.Duration {
	lim, ok := documented[path]
	if !ok || lim.Every <= 0 {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	at := l.next[path]
	if at.Before(now) {
		at = now
	}
	l.next[path] = at.Add(lim.Every)
	return at.Sub(now)
}

// acquire takes the concurrency slots a path needs; the returned func frees them.
func (l *limiter) acquire(ctx context.Context, path string) (func(), error) {
	select {
	case l.global <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	lim, ok := documented[path]
	if !ok || lim.Concurrent <= 0 {
		return func() { <-l.global }, nil
	}
	select {
	case l.slot(path, lim.Concurrent) <- struct{}{}:
		return func() { <-l.slot(path, lim.Concurrent); <-l.global }, nil
	case <-ctx.Done():
		<-l.global
		return nil, ctx.Err()
	}
}

func (l *limiter) slot(path string, n int) chan struct{} {
	l.mu.Lock()
	defer l.mu.Unlock()
	if c, ok := l.slots[path]; ok {
		return c
	}
	c := make(chan struct{}, n)
	l.slots[path] = c
	return c
}
