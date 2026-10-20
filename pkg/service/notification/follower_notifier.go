package notification

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
)

// followerDeliveryMaxAttempts bounds retries for a single follower/product
// digest before it's given up on as permanently "dead" (still visible for
// inspection, just no longer retried). At followerDeliveryRetryEvery = 5m,
// 8 attempts spans ~40 minutes — long enough to ride out a token that's
// mid-rotation or a transient FCM/Firestore hiccup, short enough that a truly
// dead device stops being retried the same day.
const (
	followerDeliveryMaxAttempts = 8
	followerDeliveryRetryEvery  = 5 * time.Minute
)

// FollowerNotifier sends a "new product" push to a shop's followers when the
// shop adds a product — at most one digest per (shop, follower, calendar day).
//
// Delivery is durable, not fire-and-forget: each qualifying follower gets a
// row in follower_notification_deliveries (migration 000045) before any send
// is attempted. NotifyNewProduct makes one immediate attempt per row (fast
// path — most deliveries complete right away, same latency as before); any
// row that fails stays "pending" and is retried by a background sweep
// (RunRetrySweep, ticked every followerDeliveryRetryEvery) until it succeeds
// or exhausts followerDeliveryMaxAttempts. This is what makes a transient
// failure (a follower's FCM token mid-rotation, a brief Firestore blip) a
// retry instead of a silent, unrepeatable miss for the rest of the day.
//
// It is still self-contained and BEST-EFFORT from the caller's point of view:
// it owns its own DB handle and FCM sender, never returns an error, and every
// failure is logged, not propagated. Callers invoke NotifyNewProduct
// fire-and-forget (a goroutine) right after a successful product save.
type FollowerNotifier struct {
	db  *gorm.DB
	fcm PushSender
	// cs resolves bare object keys (what product uploads store) to absolute
	// CDN URLs. Optional — a nil service just falls back to env config.
	cs cloud.CloudService
}

// NewFollowerNotifier builds a notifier from a GORM handle (followers + the
// delivery outbox table), any PushSender (FCM delivery), and the object
// storage service used to turn stored image keys into absolute URLs. cs may be
// nil, in which case image resolution falls back to environment config.
//
// As a side effect it starts the retry-sweep ticker in the background for the
// life of the process — safe because the API server runs as a single
// always-on replica (see cmd/api/main.go's onboarding-nudge ticker for the
// same reasoning), so there's no double-fire risk across replicas.
func NewFollowerNotifier(db *gorm.DB, fcm PushSender, cs cloud.CloudService) *FollowerNotifier {
	n := &FollowerNotifier{db: db, fcm: fcm, cs: cs}
	if db != nil && fcm != nil {
		go n.runRetryTicker(context.Background())
	}
	return n
}

// followerDeliveryRow is the subset of follower_notification_deliveries
// columns needed to (re)attempt a send.
type followerDeliveryRow struct {
	ID         string
	ShopID     string
	FollowerID string
	Title      string
	Body       string
	ImageURL   string
	Attempts   int
}

// NotifyNewProduct notifies the shop's followers about a newly added product.
// Flow (all best-effort):
//  1. Load the shop's followers — if none, do nothing.
//  2. Enqueue one durable row per follower for today (ON CONFLICT DO NOTHING
//     on (shop_id, notify_date, follower_id)) — a follower already queued
//     today for this shop is skipped, this is the once-per-follower-per-day
//     content guard.
//  3. Build the push from the product name + first image (+ shop name), and
//     make one immediate delivery attempt per newly-queued follower. Any
//     failure is recorded on its row and left for the retry sweep — it is
//     NOT lost.
//
// Safe to run in a goroutine with a detached context.
func (n *FollowerNotifier) NotifyNewProduct(ctx context.Context, shopID, productName string, imageURLs []string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WARN [FollowerNotifier]: recovered from panic for shop %s: %v", shopID, r)
		}
	}()
	if n == nil || n.db == nil || n.fcm == nil || strings.TrimSpace(shopID) == "" {
		return
	}

	// 1. Followers first — no followers, nothing to do (and nothing queued).
	var followerIDs []string
	if err := n.db.WithContext(ctx).
		Model(&domain.ShopSocial{}).
		Where("shop_id = ? AND is_follower = ?", shopID, true).
		Distinct().
		Pluck("user_id", &followerIDs).Error; err != nil {
		log.Printf("WARN [FollowerNotifier]: load followers for shop %s: %v", shopID, err)
		return
	}
	if len(followerIDs) == 0 {
		log.Printf("INFO [FollowerNotifier]: shop %s has no followers, skipping push", shopID)
		return
	}

	// 2. Build content once, then enqueue a durable row per follower.
	shopName := n.shopName(ctx, shopID)
	title, body := n.buildMessage(shopName, productName)
	imageURL := ""
	if stored := followerFirstNonBlank(imageURLs); stored != "" {
		if img := resolveFollowerImageURL(n.cs, stored); img != "" {
			imageURL = img
		} else {
			// Don't fail the push — but say so, since a silently image-less
			// notification is exactly the symptom this resolution bug caused.
			log.Printf("WARN [FollowerNotifier]: could not resolve image %q to an absolute URL for shop %s; sending without image", stored, shopID)
		}
	}

	rows, alreadyQueued, err := n.enqueueDeliveries(ctx, shopID, followerIDs, title, body, imageURL)
	if err != nil {
		log.Printf("WARN [FollowerNotifier]: enqueue deliveries for shop %s: %v", shopID, err)
		return
	}
	if alreadyQueued > 0 {
		log.Printf("INFO [FollowerNotifier]: shop %s — %d follower(s) already queued today, skipping", shopID, alreadyQueued)
	}
	if len(rows) == 0 {
		return // every follower already queued today
	}

	// 3. Immediate best-effort attempt. Failures stay "pending" in the row
	//    they were just written to — RunRetrySweep picks them up later.
	data := map[string]string{
		"event_type":   "new_product",
		"shop_id":      shopID,
		"product_name": strings.TrimSpace(productName),
	}
	if imageURL != "" {
		data["image_url"] = imageURL
	}

	sent := 0
	for _, row := range rows {
		if sendErr := n.deliverToFollower(ctx, row.FollowerID, title, body, data); sendErr != nil {
			n.recordDeliveryResult(ctx, row.ID, false, sendErr.Error())
			log.Printf("WARN [FollowerNotifier]: send to follower %s of shop %s failed (queued for retry): %v", row.FollowerID, shopID, sendErr)
			continue
		}
		n.recordDeliveryResult(ctx, row.ID, true, "")
		sent++
	}
	log.Printf("INFO [FollowerNotifier]: shop %s new-product push delivered to %d/%d followers on first attempt (failures will retry)",
		shopID, sent, len(rows))
}

// enqueueDeliveries writes one row per follower for today, skipping any
// follower already queued for this shop today. Returns only the newly created
// rows (the ones that need a send attempt) plus a count of how many were
// skipped as already-queued.
func (n *FollowerNotifier) enqueueDeliveries(ctx context.Context, shopID string, followerIDs []string, title, body, imageURL string) ([]followerDeliveryRow, int, error) {
	placeholders := make([]string, 0, len(followerIDs))
	args := make([]interface{}, 0, len(followerIDs)*6)
	total := 0
	for _, uid := range followerIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		total++
		placeholders = append(placeholders, "(?, ?, CURRENT_DATE, ?, ?, ?, ?)")
		args = append(args, domain.NewID(domain.PrefixFollowerNotifDelivery), shopID, uid, title, body, imageURL)
	}
	if len(placeholders) == 0 {
		return nil, 0, nil
	}

	query := `INSERT INTO follower_notification_deliveries
		(id, shop_id, notify_date, follower_id, title, body, image_url)
		VALUES ` + strings.Join(placeholders, ",") + `
		ON CONFLICT (shop_id, notify_date, follower_id) DO NOTHING
		RETURNING id, follower_id`

	var rows []followerDeliveryRow
	if err := n.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total - len(rows), nil
}

// recordDeliveryResult persists the outcome of one send attempt. A failure
// increments attempts and, once followerDeliveryMaxAttempts is reached, marks
// the row "dead" so the sweep stops retrying it; short of that it stays
// "pending" for the next sweep.
func (n *FollowerNotifier) recordDeliveryResult(ctx context.Context, id string, sent bool, errMsg string) {
	if sent {
		if err := n.db.WithContext(ctx).Exec(
			`UPDATE follower_notification_deliveries
			 SET status = 'sent', attempts = attempts + 1, last_error = '', updated_at = NOW()
			 WHERE id = ?`, id).Error; err != nil {
			log.Printf("WARN [FollowerNotifier]: record delivery success for %s: %v", id, err)
		}
		return
	}
	if err := n.db.WithContext(ctx).Exec(
		`UPDATE follower_notification_deliveries
		 SET attempts = attempts + 1,
		     last_error = ?,
		     status = CASE WHEN attempts + 1 >= ? THEN 'dead' ELSE 'pending' END,
		     updated_at = NOW()
		 WHERE id = ?`, errMsg, followerDeliveryMaxAttempts, id).Error; err != nil {
		log.Printf("WARN [FollowerNotifier]: record delivery failure for %s: %v", id, err)
	}
}

// deliverToFollower tries Postgres-registered device tokens first (works even
// if Firestore is unreachable, stale, or was never synced for this follower),
// falling back to the live Firestore lookup that SendToOwnerViaFirestore
// performs — the same dual-path pattern already used for admin-triggered
// pushes (see notificationUseCase.SendPushNotification). Trying both paths
// means a gap in either one alone (a stale Firestore doc, or a follower who
// never hit the Postgres-registering endpoint) doesn't cost a delivery.
func (n *FollowerNotifier) deliverToFollower(ctx context.Context, followerID, title, body string, data map[string]string) error {
	if tokens, tokErr := n.activeDeviceTokens(ctx, followerID); tokErr == nil && len(tokens) > 0 {
		if sendErr := n.fcm.SendToTokens(ctx, tokens, title, body, data); sendErr == nil {
			return nil
		}
		// Fall through to Firestore regardless of why the Postgres-token send
		// failed — it's a second independent chance, not a confirmed dead end.
	}
	return n.fcm.SendToOwnerViaFirestore(ctx, "users", followerID, title, body, data)
}

// activeDeviceTokens looks up notification_device_tokens the same way
// notificationRepository.GetActiveTokensByOwner does for admin-triggered
// pushes: matching on owner_id, admin_id, OR shop_id so it's resilient to
// which column a given registration path happened to populate. owner_type is
// fixed to "user" — followers are always customers, never sellers.
func (n *FollowerNotifier) activeDeviceTokens(ctx context.Context, followerID string) ([]string, error) {
	var tokens []string
	err := n.db.WithContext(ctx).
		Table("notification_device_tokens").
		Where("(owner_id = ? OR admin_id = ? OR shop_id = ?) AND owner_type = 'user' AND is_active = true",
			followerID, followerID, followerID).
		Pluck("token", &tokens).Error
	return tokens, err
}

// buildMessage returns the notification title/body for a new-product digest.
func (n *FollowerNotifier) buildMessage(shopName, productName string) (title, body string) {
	title = "New product from a shop you follow"
	if shopName != "" {
		title = "New at " + shopName
	}
	body = strings.TrimSpace(productName)
	if body == "" {
		body = "A new product is now available. Tap to explore."
	} else {
		body += " is now available. Tap to explore."
	}
	return title, body
}

// runRetryTicker calls RunRetrySweep every followerDeliveryRetryEvery until
// ctx is cancelled (in practice, process lifetime — this notifier has no
// explicit shutdown hook, matching its existing simplicity).
func (n *FollowerNotifier) runRetryTicker(ctx context.Context) {
	ticker := time.NewTicker(followerDeliveryRetryEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n.RunRetrySweep(ctx)
		}
	}
}

// RunRetrySweep re-attempts delivery for every "pending" row (bounded to a
// batch of 200, oldest first, so one huge backlog can't starve newer shops).
// Tokens are re-resolved fresh on every attempt — not the snapshot from when
// the row was enqueued — so a follower who registers a device (or whose token
// rotates) between the original attempt and now is picked up automatically.
func (n *FollowerNotifier) RunRetrySweep(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WARN [FollowerNotifier retry]: recovered from panic: %v", r)
		}
	}()
	if n == nil || n.db == nil || n.fcm == nil {
		return
	}

	var due []followerDeliveryRow
	if err := n.db.WithContext(ctx).
		Table("follower_notification_deliveries").
		Where("status = 'pending'").
		Order("updated_at ASC").
		Limit(200).
		Find(&due).Error; err != nil {
		log.Printf("WARN [FollowerNotifier retry]: fetch due rows: %v", err)
		return
	}
	if len(due) == 0 {
		return
	}

	sent, stillPending := 0, 0
	for _, row := range due {
		data := map[string]string{"event_type": "new_product", "shop_id": row.ShopID}
		if row.ImageURL != "" {
			data["image_url"] = row.ImageURL
		}
		if sendErr := n.deliverToFollower(ctx, row.FollowerID, row.Title, row.Body, data); sendErr != nil {
			n.recordDeliveryResult(ctx, row.ID, false, sendErr.Error())
			stillPending++
			continue
		}
		n.recordDeliveryResult(ctx, row.ID, true, "")
		sent++
	}
	log.Printf("INFO [FollowerNotifier retry]: swept %d due row(s) — %d sent, %d still pending/dead", len(due), sent, stillPending)
}

// shopName looks up the shop's display name for the notification title.
func (n *FollowerNotifier) shopName(ctx context.Context, shopID string) string {
	var name string
	_ = n.db.WithContext(ctx).
		Table("shop_details").
		Select("shop_name").
		Where("id = ?", shopID).
		Scan(&name).Error
	return strings.TrimSpace(name)
}

func followerFirstNonBlank(ss []string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

// ResolvePushImageURL absolutises a stored image value for any FCM push, not
// just the follower one. It is the exported entry point to the resolution rules
// documented on resolveFollowerImageURL below; callers outside this package
// (e.g. the shop-launch announcement) use it so every push resolves images the
// same way.
func ResolvePushImageURL(cs cloud.CloudService, path string) string {
	return resolveFollowerImageURL(cs, path)
}

// resolveFollowerImageURL turns a stored product-image value into an absolute
// URL so it renders in the push. It mirrors cloud.ResolveURL's three cases,
// then absolutises the result — FCM silently drops an image it cannot fetch,
// so a relative or wrong-host URL shows up as a notification with no picture.
//
//  1. Already absolute → unchanged.
//  2. "uploads/…" → served by the API host (StaticFS), the same rule the
//     customer app's resolveImageUrl and normalizePublicImageURL use. NOT the
//     object-storage base.
//  3. Anything else is a bare object key (what SaveBytes/SaveFile return, e.g.
//     "products/abc.jpg") → the object-storage/CDN base. This is the case
//     seller product uploads actually produce.
//
// Returns "" when a key cannot be made absolute, so the caller can log it
// rather than handing FCM a URL that will 404.
func resolveFollowerImageURL(cs cloud.CloudService, path string) string {
	p := strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
	if p == "" {
		return ""
	}
	if isAbsoluteURL(p) {
		return p
	}

	if strings.HasPrefix(p, "uploads/") || strings.HasPrefix(p, "/uploads/") {
		base := followerFirstNonBlank([]string{
			os.Getenv("NOTIFICATION_PUBLIC_BASE_URL"),
			os.Getenv("PUBLIC_BASE_URL"),
			os.Getenv("API_BASE_URL"),
			os.Getenv("APP_BASE_URL"),
		})
		if base == "" {
			base = "https://api.locazar.com" // production API host serving /uploads
		}
		return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(p, "/")
	}

	// Bare object key: resolve against object storage, exactly as
	// cloud.ResolveURL does for every other surface that renders these images.
	if cs != nil {
		if u := strings.TrimSpace(cs.PublicURL(p)); isAbsoluteURL(u) {
			return u
		}
	}
	// No storage service wired in (or it is the no-op): fall back to the same
	// env var the object storage service is configured from.
	if base := strings.TrimSpace(os.Getenv("S3_PUBLIC_BASE_URL")); base != "" {
		return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(p, "/")
	}
	return ""
}

func isAbsoluteURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
