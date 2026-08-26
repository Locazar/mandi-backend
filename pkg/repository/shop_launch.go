package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// GetShopAnnouncementTarget loads the fields a shop-launch push needs: the
// display name, the pinned coordinates that define the radius centre, and the
// stored image key. Returns gorm.ErrRecordNotFound semantics via an empty ID so
// the caller can produce a clean 404 rather than sending to (0,0).
func (r *notificationRepository) GetShopAnnouncementTarget(ctx context.Context, shopID string) (domain.ShopAnnouncementTarget, error) {
	var target domain.ShopAnnouncementTarget
	err := r.db.WithContext(ctx).
		Table("shop_details").
		Select("id, shop_name, city, latitude, longitude, shop_image_url, shop_status").
		Where("id = ? AND deleted_at IS NULL", shopID).
		Limit(1).
		Scan(&target).Error
	return target, err
}

// GetCustomerDevicesInRadius returns one row per reachable device belonging to a
// customer whose saved address falls within radiusKm of (lat, lng).
//
// Customer location only exists on `addresses` — the users table has no
// coordinates — so a customer who has never saved an address is unreachable by
// this feature no matter how close they live. A customer with several addresses
// in range must still be counted once, hence the DISTINCT on (user_id, token).
//
// Distance is the same haversine expression used by the shop search and
// conflict queries, with least(1.0, …) guarding acos against floating-point
// drift above 1 for near-zero distances.
func (r *notificationRepository) GetCustomerDevicesInRadius(ctx context.Context, lat, lng, radiusKm float64) ([]domain.CustomerDevice, error) {
	const query = `
		SELECT DISTINCT a.user_id AS user_id, t.token AS token
		FROM addresses a
		JOIN users u
			ON u.id = a.user_id
		JOIN notification_device_tokens t
			ON t.owner_id = a.user_id
			AND t.owner_type = 'user'
			AND t.is_active = true
		WHERE a.deleted_at IS NULL
			AND u.deleted_at IS NULL
			AND u.block_status = false
			AND a.latitude IS NOT NULL
			AND a.longitude IS NOT NULL
			AND (6371 * acos(least(1.0,
				cos(radians($1)) * cos(radians(a.latitude)) *
				cos(radians(a.longitude) - radians($2)) +
				sin(radians($1)) * sin(radians(a.latitude))
			))) <= $3`

	var devices []domain.CustomerDevice
	err := r.db.WithContext(ctx).Raw(query, lat, lng, radiusKm).Scan(&devices).Error
	return devices, err
}

// SaveNotificationsBatch inserts many notification rows in one round trip so a
// wide announcement does not turn into thousands of individual INSERTs. Chunked
// because Postgres caps a statement at 65535 bind parameters and each row here
// binds ~20.
func (r *notificationRepository) SaveNotificationsBatch(ctx context.Context, notifications []domain.Notification) error {
	if len(notifications) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(notifications, 500).Error
}
