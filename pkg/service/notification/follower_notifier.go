package notification

import (
	"context"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
)

// FollowerNotifier sends a "new product" push to a shop's followers when the
// shop adds a product — at most once per shop per calendar day.
//
// It is deliberately self-contained and BEST-EFFORT: it owns its own DB handle
// and FCM sender, never returns an error to the caller, and every failure is
// swallowed (logged). Callers invoke it fire-and-forget (a goroutine) right
// after a successful product save, so it can never delay, fail, or otherwise
// affect the existing product-add flow.
type FollowerNotifier struct {
	db  *gorm.DB
	fcm PushSender
	// cs resolves bare object keys (what product uploads store) to absolute
	// CDN URLs. Optional — a nil service just falls back to env config.
	cs cloud.CloudService
}

// NewFollowerNotifier builds a notifier from a GORM handle (followers + the
// once-per-day dedup table), any PushSender (FCM delivery), and the object
// storage service used to turn stored image keys into absolute URLs. cs may be
// nil, in which case image resolution falls back to environment config.
func NewFollowerNotifier(db *gorm.DB, fcm PushSender, cs cloud.CloudService) *FollowerNotifier {
	return &FollowerNotifier{db: db, fcm: fcm, cs: cs}
}

// NotifyNewProduct notifies the shop's followers about a newly added product.
// Flow (all best-effort):
//  1. Load the shop's followers — if none, do nothing.
//  2. Atomically claim today's slot (INSERT ... ON CONFLICT DO NOTHING); if a
//     product already notified today, skip — this is the once-per-day guard.
//  3. Build the push from the product name + first image (+ shop name), and
//     deliver to each follower's registered devices.
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

	// 1. Followers first — no followers, nothing to do (and no daily slot spent).
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
		return
	}

	// 2. Once-per-day guard: atomically claim today's slot. RowsAffected == 0
	//    means a product already triggered the follower push today → skip.
	claim := n.db.WithContext(ctx).Exec(
		`INSERT INTO shop_new_product_notifications (shop_id, notify_date)
		 VALUES (?, CURRENT_DATE)
		 ON CONFLICT (shop_id, notify_date) DO NOTHING`,
		shopID,
	)
	if claim.Error != nil {
		log.Printf("WARN [FollowerNotifier]: claim daily slot for shop %s: %v", shopID, claim.Error)
		return
	}
	if claim.RowsAffected == 0 {
		return // already notified today
	}

	// 3. Build and deliver.
	shopName := n.shopName(ctx, shopID)
	title := "New product from a shop you follow"
	if shopName != "" {
		title = "New at " + shopName
	}
	body := strings.TrimSpace(productName)
	if body == "" {
		body = "A new product is now available. Tap to explore."
	} else {
		body += " is now available. Tap to explore."
	}

	data := map[string]string{
		"event_type":   "new_product",
		"shop_id":      shopID,
		"product_name": strings.TrimSpace(productName),
	}
	if stored := followerFirstNonBlank(imageURLs); stored != "" {
		if img := resolveFollowerImageURL(n.cs, stored); img != "" {
			data["image_url"] = img
		} else {
			// Don't fail the push — but say so, since a silently image-less
			// notification is exactly the symptom this resolution bug caused.
			log.Printf("WARN [FollowerNotifier]: could not resolve image %q to an absolute URL for shop %s; sending without image", stored, shopID)
		}
	}

	sent := 0
	for _, uid := range followerIDs {
		if strings.TrimSpace(uid) == "" {
			continue
		}
		// A follower with no active device (logged out / no token) simply returns
		// a no-active-tokens error — expected, not a failure. Keep going.
		if err := n.fcm.SendToOwnerViaFirestore(ctx, "users", uid, title, body, data); err != nil {
			continue
		}
		sent++
	}
	log.Printf("INFO [FollowerNotifier]: shop %s new-product push delivered to %d/%d followers",
		shopID, sent, len(followerIDs))
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
