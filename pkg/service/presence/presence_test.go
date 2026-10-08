package presence

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// The throttle is what keeps this off the hot path. Verify it collapses a
// burst into one write and reopens after the window.
func TestThrottleCollapsesBurst(t *testing.T) {
	tr := &tracker{seen: make(map[string]time.Time)}

	// First call claims the slot.
	if !tr.claims("u1", time.Now()) {
		t.Fatal("first touch should write")
	}
	// Everything inside the window is suppressed.
	for i := 0; i < 50; i++ {
		if tr.claims("u1", time.Now()) {
			t.Fatalf("touch %d inside the window should NOT write", i)
		}
	}
	// A different user is independent.
	if !tr.claims("u2", time.Now()) {
		t.Error("a different user should write")
	}
	// After the window it writes again.
	tr.mu.Lock()
	tr.seen["u1"] = time.Now().Add(-throttle - time.Second)
	tr.mu.Unlock()
	if !tr.claims("u1", time.Now()) {
		t.Error("after the throttle window it should write again")
	}
}

// Two concurrent requests from one user must not both decide to write.
func TestConcurrentTouchesClaimOnce(t *testing.T) {
	tr := &tracker{seen: make(map[string]time.Time)}
	var wg sync.WaitGroup
	var mu sync.Mutex
	writes := 0
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if tr.claims("same", time.Now()) {
				mu.Lock()
				writes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if writes != 1 {
		t.Errorf("expected exactly 1 write from 100 concurrent touches, got %d", writes)
	}
}

// newTestTracker installs a tracker that records writes instead of performing
// them, and returns a function to read what was written.
func newTestTracker(t *testing.T) func() []string {
	t.Helper()
	var mu sync.Mutex
	var written []string
	tr := &tracker{seen: make(map[string]time.Time)}
	tr.write = func(userID string) {
		mu.Lock()
		written = append(written, userID)
		mu.Unlock()
	}
	prev := shared
	shared = tr
	t.Cleanup(func() { shared = prev })
	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), written...)
	}
}

// The middleware must see a userId that the ROUTE GROUP's auth middleware sets,
// which only works because it reads the context after c.Next(). This mounts it
// the same way server.go does: Use() before the routes are registered.
func TestMiddlewareStampsUserSetByGroupAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	writes := newTestTracker(t)

	e := gin.New()
	e.Use(Middleware()) // before route registration, as in NewServerHTTP
	api := e.Group("/api")
	api.Use(func(c *gin.Context) { c.Set("userId", "seller-1"); c.Next() }) // stand-in for auth
	api.GET("/alerts", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/alerts", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := writes(); len(got) != 1 || got[0] != "seller-1" {
		t.Fatalf("writes = %v, want [seller-1]", got)
	}
}

// A failed request is not evidence that anyone is using the app.
func TestMiddlewareSkipsUnauthenticatedAndFailedRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name   string
		handle gin.HandlerFunc
	}{
		{"no userId at all", func(c *gin.Context) { c.Status(http.StatusOK) }},
		{"401 with userId", func(c *gin.Context) {
			c.Set("userId", "seller-1")
			c.Status(http.StatusUnauthorized)
		}},
		{"500 with userId", func(c *gin.Context) {
			c.Set("userId", "seller-1")
			c.Status(http.StatusInternalServerError)
		}},
		{"blank userId", func(c *gin.Context) {
			c.Set("userId", "")
			c.Status(http.StatusOK)
		}},
		{"non-string userId", func(c *gin.Context) {
			c.Set("userId", 42)
			c.Status(http.StatusOK)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			writes := newTestTracker(t)
			e := gin.New()
			e.Use(Middleware())
			e.GET("/x", tc.handle)
			e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))
			if got := writes(); len(got) != 0 {
				t.Fatalf("writes = %v, want none", got)
			}
		})
	}
}

// An uninitialised tracker must be an explicit no-op, not a nil panic.
func TestMiddlewareIsNoOpBeforeInit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prev := shared
	shared = nil
	t.Cleanup(func() { shared = prev })

	e := gin.New()
	e.Use(Middleware())
	e.GET("/x", func(c *gin.Context) { c.Set("userId", "s1"); c.Status(http.StatusOK) })
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

// A chatty client must cost one write per window, not one per request.
func TestMiddlewareThrottlesRepeatRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	writes := newTestTracker(t)

	e := gin.New()
	e.Use(Middleware())
	e.GET("/x", func(c *gin.Context) { c.Set("userId", "s1"); c.Status(http.StatusOK) })
	for i := 0; i < 50; i++ {
		e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))
	}
	if got := writes(); len(got) != 1 {
		t.Fatalf("writes = %d, want exactly 1", len(got))
	}
}

// The throttle map must not grow without bound: this process runs for weeks and
// every distinct caller — including customers, who never match an admins row —
// gets an entry.
func TestThrottleMapStaysBounded(t *testing.T) {
	tr := &tracker{seen: make(map[string]time.Time)}
	tr.write = func(string) {}

	// Half the entries are already outside the window, so they are evictable.
	stale := time.Now().Add(-2 * throttle)
	for i := 0; i < maxTracked; i++ {
		tr.seen[fmt.Sprintf("old-%d", i)] = stale
	}
	for i := 0; i < maxTracked; i++ {
		tr.touch(fmt.Sprintf("new-%d", i))
	}
	if len(tr.seen) > maxTracked {
		t.Fatalf("map size = %d, want <= %d", len(tr.seen), maxTracked)
	}

	// A burst of distinct users all inside the window must also stay bounded.
	for i := 0; i < maxTracked+1000; i++ {
		tr.touch(fmt.Sprintf("burst-%d", i))
	}
	if len(tr.seen) > maxTracked {
		t.Fatalf("map size after burst = %d, want <= %d", len(tr.seen), maxTracked)
	}
}

// Production guarantee: the middleware observes, it never alters. A route's
// status, headers and body must be byte-identical with and without it.
func TestMiddlewareLeavesResponseUntouched(t *testing.T) {
	gin.SetMode(gin.TestMode)

	build := func(withPresence bool) *httptest.ResponseRecorder {
		e := gin.New()
		if withPresence {
			e.Use(Middleware())
		}
		e.GET("/x", func(c *gin.Context) {
			c.Set("userId", "seller-1")
			c.Header("X-Custom", "kept")
			c.JSON(http.StatusCreated, gin.H{"ok": true})
		})
		w := httptest.NewRecorder()
		e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
		return w
	}

	writes := newTestTracker(t)
	with := build(true)
	without := build(false)

	if with.Code != without.Code {
		t.Fatalf("status differs: with=%d without=%d", with.Code, without.Code)
	}
	if with.Body.String() != without.Body.String() {
		t.Fatalf("body differs:\n with=%q\n without=%q", with.Body.String(), without.Body.String())
	}
	if with.Header().Get("X-Custom") != without.Header().Get("X-Custom") {
		t.Fatal("headers differ")
	}
	// ...and it did do its job, so this is not passing by being a no-op.
	if got := writes(); len(got) != 1 {
		t.Fatalf("writes = %v, want 1 (test would be vacuous otherwise)", got)
	}
}

// Over capacity, stamps must be dropped rather than queued: queuing would put
// liveness telemetry in competition with live traffic for the shared pool.
func TestStampDropsWritesBeyondCapacity(t *testing.T) {
	block := make(chan struct{})
	var started sync.WaitGroup

	tr := &tracker{
		seen:     make(map[string]time.Time),
		inflight: make(chan struct{}, maxInflight),
	}
	// Occupy every slot with a write that will not finish until released.
	tr.write = func(string) {
		select {
		case tr.inflight <- struct{}{}:
			started.Done()
			go func() { <-block; <-tr.inflight }()
		default:
		}
	}
	started.Add(maxInflight)
	for i := 0; i < maxInflight; i++ {
		tr.touch(fmt.Sprintf("busy-%d", i))
	}
	started.Wait()

	// The real stamp path must now refuse immediately rather than block.
	tr.write = tr.stamp
	done := make(chan struct{})
	go func() {
		tr.touch("overflow-user") // would hang forever if it queued
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stamp blocked when at capacity; it must drop the write instead")
	}
	close(block)
}

// A panic on the write goroutine is not caught by gin's recovery middleware and
// would terminate the process. This drives a real nil-database panic through
// the production write path: the test binary surviving IS the assertion.
func TestStampSurvivesPanicInWritePath(t *testing.T) {
	tr := &tracker{
		db:       nil, // any use panics
		seen:     make(map[string]time.Time),
		inflight: make(chan struct{}, maxInflight),
	}
	tr.write = tr.stamp
	tr.touch("seller-1")

	// Wait for the slot to be released, which only happens via the deferred
	// cleanup — i.e. the goroutine unwound through recover() rather than
	// crashing the process.
	deadline := time.Now().Add(2 * time.Second)
	for len(tr.inflight) > 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if len(tr.inflight) != 0 {
		t.Fatal("inflight slot never released; write goroutine did not unwind cleanly")
	}
}
