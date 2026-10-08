// Package presence records when a seller or customer last used the app.
//
// The apps already call authenticated endpoints on every launch — app-config,
// feature-flags, alerts, shop details — so liveness can be observed server-side
// with no client change and no app release. That matters because FCM tokens
// rotate only every few months, so a token's age says almost nothing about
// whether anyone is still opening the app.
//
// Three rules shape this file, because it sits on the path of EVERY
// authenticated request:
//
//  1. It never blocks the response. The write happens after the handler has
//     finished, on its own goroutine.
//  2. It never fails a request. Every error is swallowed and logged at most.
//  3. It almost never writes. An in-memory throttle means at most one UPDATE
//     per user per interval, so a chatty client costs one row write every
//     15 minutes rather than one per call.
//
// Initialised once at startup, like InitSharedFirebaseApp: uninitialised, the
// middleware is an explicit no-op rather than a nil-pointer panic.
package presence

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// throttle is how long to wait before writing the same user's timestamp again.
// Fifteen minutes keeps "active today" and "active this week" perfectly
// accurate while collapsing a whole session into a single write.
const throttle = 15 * time.Minute

// Hard limits on what this feature may cost the rest of the system. The
// database pool is shared with real request handling (240 connections, see
// pkg/db/connection.go), so liveness telemetry must never be able to queue on
// it: a push notification can wake tens of thousands of sellers at once.
const (
	// maxInflight is how many stamp writes may be in flight at once. Beyond
	// this, writes are DROPPED rather than queued — a missed timestamp is
	// invisible in the numbers, a starved connection pool is an outage.
	maxInflight = 4
	// writeTimeout bounds a single stamp, so a slow or locked database cannot
	// hold a pooled connection indefinitely.
	writeTimeout = 3 * time.Second
	// maxTracked caps the throttle map. Every distinct caller gets an entry —
	// including customers, whose ids never match an admins row — and this
	// process runs for weeks, so without a cap the map would only ever grow.
	maxTracked = 50_000
)

type tracker struct {
	db *gorm.DB

	// inflight is a counting semaphore: a slot must be acquired before a write,
	// and acquisition never blocks.
	inflight chan struct{}

	// write performs the actual persistence. A field rather than a direct call
	// so tests can observe exactly which users got written, and how often,
	// without a database.
	write func(userID string)

	mu   sync.Mutex
	seen map[string]time.Time
}

var shared *tracker

// Init wires the tracker to the database. Safe to call once at startup; calling
// it again replaces the tracker, which is only useful in tests.
func Init(db *gorm.DB) {
	t := &tracker{
		db:       db,
		seen:     make(map[string]time.Time),
		inflight: make(chan struct{}, maxInflight),
	}
	t.write = t.stamp
	shared = t
}

// Middleware stamps last_seen_at for whichever user the auth middleware
// resolved.
//
// It calls c.Next() FIRST and reads the context afterwards: "userId" is set by
// the route group's auth middleware, which runs inside this call. Reading it
// before would always find nothing.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if shared == nil {
			return
		}
		// Only count a request that actually authenticated and succeeded. A 401
		// or a 500 is not evidence that someone is using the app.
		if c.Writer.Status() >= 400 {
			return
		}
		raw, ok := c.Get("userId")
		if !ok {
			return
		}
		userID, ok := raw.(string)
		if !ok || userID == "" {
			return
		}
		shared.touch(userID)
	}
}

// claims reports whether this touch is the one that should write, and records
// the claim. Split out from touch so the throttle can be tested without a
// database: concurrent callers must see exactly one true per window.
func (t *tracker) claims(userID string, now time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if last, found := t.seen[userID]; found && now.Sub(last) < throttle {
		return false
	}
	// Claim before releasing the lock, so two concurrent requests from the same
	// user cannot both decide to write.
	t.seen[userID] = now
	t.pruneLocked(now)
	return true
}

// pruneLocked drops entries that are already past the throttle window, so
// evicting them costs nothing but a repeat write for anyone still active.
// Caller must hold t.mu. Only runs when the map is over the cap, so the usual
// path stays a single map write.
func (t *tracker) pruneLocked(now time.Time) {
	if len(t.seen) <= maxTracked {
		return
	}
	for id, last := range t.seen {
		if now.Sub(last) >= throttle {
			delete(t.seen, id)
		}
	}
	// Everything is still inside the window (a genuine burst of distinct
	// users): drop the whole map rather than grow without bound. The cost is
	// one extra write per affected user, never a wrong answer.
	if len(t.seen) > maxTracked {
		t.seen = make(map[string]time.Time, maxTracked)
	}
}

// touch records a user as seen, at most once per throttle window.
func (t *tracker) touch(userID string) {
	if !t.claims(userID, time.Now()) {
		return
	}

	t.write(userID)
}

// stamp is the production write: detached, because the response has already
// been sent and a slow or failing database must not surface to the client or
// hold the connection open.
//
// Only seller/platform accounts live in admins, so a customer's userId matches
// no row and updates nothing. That is deliberate — customers mostly browse the
// web app, where "still installed" has no meaning — and costs one no-op UPDATE
// per customer per throttle window.
func (t *tracker) stamp(userID string) {
	// Non-blocking acquire. If all slots are busy, give up immediately: the
	// same user's next request after the throttle window will try again, and
	// dropping telemetry is always preferable to competing with live traffic
	// for a pooled connection.
	select {
	case t.inflight <- struct{}{}:
	default:
		return
	}

	go func() {
		defer func() { <-t.inflight }()
		// A panic on its own goroutine is NOT caught by gin's recovery
		// middleware and would take the whole process down. Nothing here is
		// expected to panic, but liveness telemetry must not be able to stop
		// the API serving traffic.
		defer func() {
			if r := recover(); r != nil {
				log.Printf("WARN [presence]: recovered while stamping %s: %v", userID, r)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		defer cancel()

		if err := t.db.WithContext(ctx).Exec(
			`UPDATE admins SET last_seen_at = NOW() WHERE id = ?`, userID,
		).Error; err != nil {
			// Not fatal and not worth retrying — the next request after the
			// throttle window will try again.
			log.Printf("WARN [presence]: could not stamp last_seen_at for %s: %v", userID, err)
		}
	}()
}

// ── reporting ──────────────────────────────────────────────────────────────

// Stats answers "who is still using the app, and who uninstalled it".
//
// Two independent signals, deliberately reported side by side because they
// disagree in a useful way:
//
//   - AppActivity comes from last_seen_at, written by the middleware above. It
//     means "opened the app", and it is only as old as this feature.
//   - Devices comes from fcm_tokens.is_active, which the push sender has
//     always maintained: a token is deactivated when FCM reports the app was
//     uninstalled. That is authoritative but LAZY — it only updates when a
//     notification is actually sent to that device, so a seller who uninstalled
//     last week still counts as installed until something targets them.
type Stats struct {
	AppActivity ActivityCounts `json:"app_activity"`
	Devices     []DeviceCounts `json:"devices"`
	GeneratedAt time.Time      `json:"generated_at"`
}

type ActivityCounts struct {
	Active24h int64 `json:"active_24h"`
	Active7d  int64 `json:"active_7d"`
	Active30d int64 `json:"active_30d"`
	// Dormant has a timestamp but older than 30 days.
	Dormant int64 `json:"dormant"`
	// NeverSeen has no timestamp at all — either signed up before this feature
	// shipped and has not returned, or never opened the app.
	NeverSeen int64 `json:"never_seen"`
	Total     int64 `json:"total"`
}

type DeviceCounts struct {
	OwnerType   string `json:"owner_type"`
	Installed   int64  `json:"installed"`
	Uninstalled int64  `json:"uninstalled"`
}

// GetStats is read-only: two aggregate queries, no index needed. The admins
// aggregate is a parallel seq scan measured at ~15 ms over 200k rows, and this
// endpoint is admin-only and hand-triggered, so it is not on any hot path.
func GetStats(ctx context.Context) (Stats, error) {
	out := Stats{GeneratedAt: time.Now()}
	if shared == nil {
		return out, errors.New("presence not initialised")
	}
	db := shared.db.WithContext(ctx)

	// App users are the admins rows that are NOT platform/back-office users.
	// Defined as the exact complement of the platform-user listing in
	// pkg/repository/platform_user.go, which excludes blank, NULL and 'seller'
	// — so those three values are precisely the app accounts. Migration 000025
	// documents why blank is the normal value for a self-signed-up seller.
	if err := db.Raw(`
		SELECT
			COUNT(*) FILTER (WHERE last_seen_at >= NOW() - INTERVAL '24 hours') AS active24h,
			COUNT(*) FILTER (WHERE last_seen_at >= NOW() - INTERVAL '7 days')   AS active7d,
			COUNT(*) FILTER (WHERE last_seen_at >= NOW() - INTERVAL '30 days')  AS active30d,
			COUNT(*) FILTER (WHERE last_seen_at IS NOT NULL
			                   AND last_seen_at <  NOW() - INTERVAL '30 days')  AS dormant,
			COUNT(*) FILTER (WHERE last_seen_at IS NULL)                        AS never_seen,
			COUNT(*)                                                            AS total
		FROM admins
		WHERE COALESCE(role, '') IN ('', 'seller')
		  AND deleted_at IS NULL
	`).Scan(&out.AppActivity).Error; err != nil {
		return out, err
	}

	// Counted per OWNER, not per token, and "uninstalled" means "has no live
	// token left" — total distinct owners minus those with one. Filtering on
	// NOT is_active instead would count anyone who changed phones or
	// reinstalled in BOTH columns, since their old token stays deactivated,
	// which inflates churn.
	//
	// Grouped by whatever owner_type values are actually present rather than an
	// assumed list, since the column is free-form text.
	if err := db.Raw(`
		SELECT owner_type,
		       COUNT(DISTINCT owner_id) FILTER (WHERE is_active) AS installed,
		       COUNT(DISTINCT owner_id)
		         - COUNT(DISTINCT owner_id) FILTER (WHERE is_active) AS uninstalled
		FROM fcm_tokens
		GROUP BY owner_type
		ORDER BY owner_type
	`).Scan(&out.Devices).Error; err != nil {
		return out, err
	}

	return out, nil
}
