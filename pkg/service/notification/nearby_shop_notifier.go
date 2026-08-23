package notification

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// defaultNearbyShopRadiusKm is how far a "new shop nearby" push travels.
// Override per-environment with NEARBY_SHOP_RADIUS_KM.
const defaultNearbyShopRadiusKm = 1.0

// nearbyShopNotifyLimit caps one broadcast so a shop opening in a dense area
// cannot fan out unboundedly in a single request.
const nearbyShopNotifyLimit = 5000

// NearbyShopNotifier tells customers near a shop that it has just opened —
// once per shop, ever.
//
// Like FollowerNotifier this is deliberately self-contained and BEST-EFFORT:
// it owns its own DB handle and FCM sender, never returns an error, and
// swallows (logs) every failure. Callers invoke it fire-and-forget so it can
// never delay, fail, or otherwise affect the shop onboarding flow.
type NearbyShopNotifier struct {
	db  *gorm.DB
	fcm PushSender
}

// NewNearbyShopNotifier builds a notifier from a GORM handle (shop lookup,
// customer geo query, and the once-ever dedup table) and any PushSender.
func NewNearbyShopNotifier(db *gorm.DB, fcm PushSender) *NearbyShopNotifier {
	return &NearbyShopNotifier{db: db, fcm: fcm}
}

// nearbyShop is the subset of shop_details the broadcast needs.
// Columns are tagged explicitly rather than left to GORM's name inference, so
// a naming-strategy change can never silently blank a field (an unmapped
// shop_status would read as "" and suppress every broadcast).
type nearbyShop struct {
	ID           string   `gorm:"column:id"`
	ShopName     string   `gorm:"column:shop_name"`
	AddressLine1 string   `gorm:"column:address_line1"`
	AddressLine2 string   `gorm:"column:address_line2"`
	City         string   `gorm:"column:city"`
	Pincode      string   `gorm:"column:pincode"`
	ShopImageURL string   `gorm:"column:shop_image_url"`
	Latitude     *float64 `gorm:"column:latitude"`
	Longitude    *float64 `gorm:"column:longitude"`
	ShopStatus   string   `gorm:"column:shop_status"`
}

// NotifyNewShop announces a newly live shop to nearby customers.
// Flow (all best-effort):
//  1. Load the shop; skip unless it is 'active' and has coordinates. A shop in
//     any other state is invisible to customers (see ShopActivePredicate), so
//     announcing it would send them to a dead end.
//  2. Atomically claim the shop's one-time slot (INSERT ... ON CONFLICT DO
//     NOTHING); already claimed → skip. This is what makes a seller re-saving
//     onboarding details, or an approve/suspend/approve cycle, harmless.
//  3. Find customers with a saved address within the radius and push the shop
//     name, address, and image to each.
//
// Safe to run in a goroutine with a detached context.
func (n *NearbyShopNotifier) NotifyNewShop(ctx context.Context, shopID string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WARN [NearbyShopNotifier]: recovered from panic for shop %s: %v", shopID, r)
		}
	}()
	if n == nil || n.db == nil || n.fcm == nil || strings.TrimSpace(shopID) == "" {
		return
	}

	// 1. Shop must be live and locatable.
	shop, ok := n.loadShop(ctx, shopID)
	if !ok {
		return
	}
	if shop.ShopStatus != "active" {
		return // not yet approved — announcing it would link to a hidden shop
	}
	if shop.Latitude == nil || shop.Longitude == nil {
		log.Printf("INFO [NearbyShopNotifier]: shop %s has no coordinates, skipping nearby broadcast", shopID)
		return
	}

	// 2. Once-ever guard, claimed BEFORE the fan-out so two concurrent
	//    approvals cannot both broadcast.
	claim := n.db.WithContext(ctx).Exec(
		`INSERT INTO shop_launch_notifications (shop_id)
		 VALUES (?)
		 ON CONFLICT (shop_id) DO NOTHING`,
		shopID,
	)
	if claim.Error != nil {
		log.Printf("WARN [NearbyShopNotifier]: claim launch slot for shop %s: %v", shopID, claim.Error)
		return
	}
	if claim.RowsAffected == 0 {
		return // already announced
	}

	// 3. Fan out.
	customerIDs := n.nearbyCustomerIDs(ctx, *shop.Latitude, *shop.Longitude, shopID)
	if len(customerIDs) == 0 {
		log.Printf("INFO [NearbyShopNotifier]: no customers within %.2fkm of shop %s", nearbyShopRadiusKm(), shopID)
		return
	}

	title := "New shop near you"
	if name := strings.TrimSpace(shop.ShopName); name != "" {
		title = name + " just opened near you"
	}
	body := nearbyShopBody(shop)

	data := map[string]string{
		"event_type":   "new_shop_nearby",
		"shop_id":      shop.ID,
		"shop_name":    strings.TrimSpace(shop.ShopName),
		"shop_address": nearbyShopAddress(shop),
	}
	if img := resolvePublicNotificationImageURL(shop.ShopImageURL); img != "" {
		data["image_url"] = img
	}

	sent := 0
	for _, uid := range customerIDs {
		if strings.TrimSpace(uid) == "" {
			continue
		}
		// A customer with no active device (logged out / no token) returns a
		// no-active-tokens error — expected, not a failure. Keep going.
		if err := n.fcm.SendToOwnerViaFirestore(ctx, "users", uid, title, body, data); err != nil {
			continue
		}
		sent++
	}
	log.Printf("INFO [NearbyShopNotifier]: shop %s launch push delivered to %d/%d nearby customers",
		shopID, sent, len(customerIDs))
}

// loadShop reads the fields the broadcast needs. A missing shop returns false.
func (n *NearbyShopNotifier) loadShop(ctx context.Context, shopID string) (nearbyShop, bool) {
	var shop nearbyShop
	err := n.db.WithContext(ctx).
		Raw(`SELECT id, shop_name, address_line1, address_line2, city, pincode,
		            shop_image_url, latitude, longitude, shop_status
		     FROM shop_details
		     WHERE id = ? AND deleted_at IS NULL`, shopID).
		Scan(&shop).Error
	if err != nil {
		log.Printf("WARN [NearbyShopNotifier]: load shop %s: %v", shopID, err)
		return nearbyShop{}, false
	}
	if strings.TrimSpace(shop.ID) == "" {
		return nearbyShop{}, false
	}
	return shop, true
}

// nearbyCustomerIDs returns the distinct ids of active customers with at least
// one saved address inside the radius.
//
// Customer coordinates live on `addresses` (the users table has none), reached
// through the user_addresses join. The haversine expression matches the one
// used by the shop-search queries, including the LEAST/GREATEST clamp that
// keeps float error from pushing acos outside [-1, 1] — which matters more at
// a 1km radius than at search distances.
func (n *NearbyShopNotifier) nearbyCustomerIDs(ctx context.Context, lat, lng float64, shopID string) []string {
	var ids []string
	err := n.db.WithContext(ctx).
		Raw(`SELECT DISTINCT u.id
		     FROM users u
		     JOIN user_addresses ua ON ua.user_id = u.id
		     JOIN addresses a ON a.id = ua.address_id
		     WHERE u.deleted_at IS NULL
		       AND u.block_status = FALSE
		       AND a.deleted_at IS NULL
		       AND a.latitude IS NOT NULL
		       AND a.longitude IS NOT NULL
		       AND (6371 * acos(
		             LEAST(1, GREATEST(-1,
		               cos(radians(?)) * cos(radians(a.latitude::double precision)) *
		               cos(radians(a.longitude::double precision) - radians(?)) +
		               sin(radians(?)) * sin(radians(a.latitude::double precision))
		             ))
		           )) <= ?
		     LIMIT ?`,
			lat, lng, lat, nearbyShopRadiusKm(), nearbyShopNotifyLimit).
		Scan(&ids).Error
	if err != nil {
		log.Printf("WARN [NearbyShopNotifier]: nearby customer lookup for shop %s: %v", shopID, err)
		return nil
	}
	return ids
}

// nearbyShopRadiusKm reads the broadcast radius, defaulting to 1km. A garbage
// or non-positive override falls back to the default rather than disabling the
// feature or blasting the whole country.
func nearbyShopRadiusKm() float64 {
	raw := strings.TrimSpace(os.Getenv("NEARBY_SHOP_RADIUS_KM"))
	if raw == "" {
		return defaultNearbyShopRadiusKm
	}
	km, err := strconv.ParseFloat(raw, 64)
	if err != nil || km <= 0 {
		log.Printf("WARN [NearbyShopNotifier]: invalid NEARBY_SHOP_RADIUS_KM %q, using %.2f", raw, defaultNearbyShopRadiusKm)
		return defaultNearbyShopRadiusKm
	}
	return km
}

// nearbyShopAddress renders the shop's address as one human-readable line,
// skipping the blank parts so it never contains ", ," runs.
func nearbyShopAddress(shop nearbyShop) string {
	parts := make([]string, 0, 4)
	for _, p := range []string{shop.AddressLine1, shop.AddressLine2, shop.City, shop.Pincode} {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, ", ")
}

// nearbyShopBody is the push body: the address when there is one, otherwise a
// generic invitation, so the notification is never an empty line.
func nearbyShopBody(shop nearbyShop) string {
	if addr := nearbyShopAddress(shop); addr != "" {
		return fmt.Sprintf("%s. Tap to explore.", addr)
	}
	return "A new shop just opened near you. Tap to explore."
}
