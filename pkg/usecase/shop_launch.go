package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	notificationSvc "github.com/rohit221990/mandi-backend/pkg/service/notification"
)

// fcmMulticastLimit is FCM's hard cap on tokens per multicast call. Radius
// audiences routinely exceed it, so every send is chunked.
const fcmMulticastLimit = 500

// ErrAnnouncementShopNotFound is returned when the announcement targets a shop
// id that does not exist (or is soft-deleted). Distinct from the existing
// ErrShopNotFound, which means "this seller has no shop registered".
var ErrAnnouncementShopNotFound = errors.New("shop not found")

// ErrShopHasNoLocation is returned when the shop has never had its position
// pinned. Sending anyway would centre the radius on (0, 0) and reach nobody.
var ErrShopHasNoLocation = errors.New("shop has no pinned location")

// PreviewShopLaunchAudience resolves how many customers and devices a send would
// reach, without delivering anything. The admin portal calls this as the radius
// slider moves, so the admin sees the blast size before committing.
func (uc *notificationUseCase) PreviewShopLaunchAudience(ctx context.Context, shopID string, radiusKm float64) (domain.RadiusAnnouncementResult, error) {
	shop, err := uc.loadAnnouncementShop(ctx, shopID)
	if err != nil {
		return domain.RadiusAnnouncementResult{}, err
	}

	devices, err := uc.notificationRepo.GetCustomerDevicesInRadius(ctx, shop.Latitude, shop.Longitude, radiusKm)
	if err != nil {
		return domain.RadiusAnnouncementResult{}, fmt.Errorf("resolve audience for shop %s: %w", shopID, err)
	}

	userIDs, tokens := splitDevices(devices)
	return domain.RadiusAnnouncementResult{
		ShopID:        shop.ID,
		ShopName:      shop.ShopName,
		RadiusKm:      radiusKm,
		CustomerCount: len(userIDs),
		DeviceCount:   len(tokens),
		ImageURL:      uc.announcementImageURL(shop, ""),
	}, nil
}

// SendShopLaunchAnnouncement pushes a "this shop just opened" notification to
// every customer with a saved address within req.RadiusKm of the shop, and
// records one in-app notification row per customer so the announcement survives
// the push banner being dismissed.
//
// An empty audience is a success, not an error: a genuinely new area with no
// registered customers nearby is a normal outcome, and failing the request would
// invite the admin to retry a send that can never do anything.
func (uc *notificationUseCase) SendShopLaunchAnnouncement(ctx context.Context, adminID string, req request.ShopLaunchAnnouncement) (domain.RadiusAnnouncementResult, error) {
	shop, err := uc.loadAnnouncementShop(ctx, req.ShopID)
	if err != nil {
		return domain.RadiusAnnouncementResult{}, err
	}

	devices, err := uc.notificationRepo.GetCustomerDevicesInRadius(ctx, shop.Latitude, shop.Longitude, req.RadiusKm)
	if err != nil {
		return domain.RadiusAnnouncementResult{}, fmt.Errorf("resolve audience for shop %s: %w", req.ShopID, err)
	}

	spec := radiusSend{
		adminID:  adminID,
		lat:      shop.Latitude,
		lng:      shop.Longitude,
		radiusKm: req.RadiusKm,
		title:    req.Title,
		body:     req.Body,
		imageURL: uc.announcementImageURL(shop, req.ImageURL),
		kind:     "shop_launch",
		shopID:   shop.ID,
		shopName: shop.ShopName,
		// Lets the customer app deep-link straight into the shop that opened.
		extraData: map[string]string{"screen": "shop", "shop_name": shop.ShopName},
	}
	return uc.deliverRadiusAnnouncement(ctx, spec, devices)
}

// SendRadiusAnnouncement is the raw-coordinate version: same targeting and
// delivery, but centred on a point the admin supplies rather than on a shop.
func (uc *notificationUseCase) SendRadiusAnnouncement(ctx context.Context, adminID string, req request.RadiusAnnouncement) (domain.RadiusAnnouncementResult, error) {
	devices, err := uc.notificationRepo.GetCustomerDevicesInRadius(ctx, req.Latitude, req.Longitude, req.RadiusKm)
	if err != nil {
		return domain.RadiusAnnouncementResult{}, fmt.Errorf("resolve audience at %f,%f: %w", req.Latitude, req.Longitude, err)
	}
	spec := radiusSend{
		adminID:  adminID,
		lat:      req.Latitude,
		lng:      req.Longitude,
		radiusKm: req.RadiusKm,
		title:    req.Title,
		body:     req.Body,
		imageURL: notificationSvc.ResolvePushImageURL(uc.cloudService, strings.TrimSpace(req.ImageURL)),
		kind:     "geo_announcement",
	}
	return uc.deliverRadiusAnnouncement(ctx, spec, devices)
}

// radiusSend is everything one radius-targeted announcement needs, so the shop
// and raw-coordinate entry points share a single delivery path.
type radiusSend struct {
	adminID   string
	lat, lng  float64
	radiusKm  float64
	title     string
	body      string
	imageURL  string
	kind      string // notification `type` column and the app's payload switch
	shopID    string // empty for a raw-coordinate push
	shopName  string
	extraData map[string]string
}

// deliverRadiusAnnouncement pushes to every resolved device and records the
// in-app copies. Shared by both entry points; see SendShopLaunchAnnouncement for
// why an empty audience is a success rather than an error.
func (uc *notificationUseCase) deliverRadiusAnnouncement(ctx context.Context, spec radiusSend, devices []domain.CustomerDevice) (domain.RadiusAnnouncementResult, error) {
	userIDs, tokens := splitDevices(devices)
	result := domain.RadiusAnnouncementResult{
		ShopID:        spec.shopID,
		ShopName:      spec.shopName,
		RadiusKm:      spec.radiusKm,
		CustomerCount: len(userIDs),
		DeviceCount:   len(tokens),
		ImageURL:      spec.imageURL,
	}
	if len(tokens) == 0 {
		log.Printf("INFO [%s]: no reachable customers within %.1fkm of %f,%f — nothing sent",
			spec.kind, spec.radiusKm, spec.lat, spec.lng)
		return result, nil
	}

	data := map[string]string{
		"type":      spec.kind,
		"radius_km": strconv.FormatFloat(spec.radiusKm, 'f', -1, 64),
	}
	for k, v := range spec.extraData {
		data[k] = v
	}
	if spec.shopID != "" {
		data["shop_id"] = spec.shopID
	}
	if spec.imageURL != "" {
		data["image_url"] = spec.imageURL
	} else {
		// Worth a line in the log: a launch push with no storefront picture is
		// the weakest version of this notification, and the cause is always
		// either a shop with no photo or an unresolvable image key.
		log.Printf("WARN [%s]: no usable image; sending without one", spec.kind)
	}

	// Deliver in FCM-sized chunks. One bad chunk must not abandon the rest of
	// the audience, so failures are counted and reported, never fatal.
	delivered := 0
	var lastErr error
	for start := 0; start < len(tokens); start += fcmMulticastLimit {
		end := start + fcmMulticastLimit
		if end > len(tokens) {
			end = len(tokens)
		}
		chunk := tokens[start:end]
		if err := uc.fcmPush.SendToTokens(ctx, chunk, spec.title, spec.body, data); err != nil {
			// Every token in the chunk being unregistered is expected churn
			// (uninstalls), not a system failure worth surfacing.
			if !errors.Is(err, notificationSvc.ErrAllTokensUnreachable) {
				lastErr = err
			}
			log.Printf("WARN [%s]: chunk %d-%d failed: %v", spec.kind, start, end, err)
			continue
		}
		delivered += len(chunk)
	}
	result.DeliveredDevices = delivered

	// In-app copies are best-effort: the push has already gone out, and failing
	// the request here would tell the admin nothing was sent when it was.
	if err := uc.recordRadiusNotifications(ctx, spec, userIDs); err != nil {
		log.Printf("WARN [%s]: push delivered but in-app records failed: %v", spec.kind, err)
	}

	if delivered == 0 && lastErr != nil {
		return result, fmt.Errorf("%s push reached no device: %w", spec.kind, lastErr)
	}
	log.Printf("INFO [%s]: delivered to %d/%d devices across %d customers within %.1fkm",
		spec.kind, delivered, len(tokens), len(userIDs), spec.radiusKm)
	return result, nil
}

// recordRadiusNotifications writes the in-app copy each targeted customer sees
// in their notification list.
func (uc *notificationUseCase) recordRadiusNotifications(ctx context.Context, spec radiusSend, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}
	// The meta blob carries what the in-app row needs to render and deep-link;
	// the column is free-form TEXT, and JSON is what the other writers use.
	meta := fmt.Sprintf(`{"type":%q,"shop_id":%q,"image_url":%q}`, spec.kind, spec.shopID, spec.imageURL)

	// domain.Notification's BeforeCreate hook assigns each id.
	now := time.Now().UTC().Format(time.RFC3339)
	rows := make([]domain.Notification, 0, len(userIDs))
	for _, uid := range userIDs {
		rows = append(rows, domain.Notification{
			SenderType:   domain.UserTypeAdmin,
			ReceiverType: domain.UserTypeUser,
			Type:         spec.kind,
			SenderID:     spec.adminID,
			ReceiverID:   uid,
			UserID:       uid,
			AdminID:      spec.adminID,
			ShopID:       spec.shopID,
			Title:        spec.title,
			Message:      spec.body,
			Body:         spec.body,
			// category_id is NOT NULL with no default; this announcement has no
			// category, so it stores the empty string rather than a fake id.
			CategoryID:           "",
			NotificationMetaData: meta,
			Status:               domain.NotificationStatusSent,
			CreatedAt:            now,
			UpdatedAt:            now,
		})
	}
	return uc.notificationRepo.SaveNotificationsBatch(ctx, rows)
}

// loadAnnouncementShop fetches the shop and rejects the two states that make an
// announcement meaningless: it does not exist, or it was never pinned to a map.
func (uc *notificationUseCase) loadAnnouncementShop(ctx context.Context, shopID string) (domain.ShopAnnouncementTarget, error) {
	shop, err := uc.notificationRepo.GetShopAnnouncementTarget(ctx, shopID)
	if err != nil {
		return shop, fmt.Errorf("load shop %s: %w", shopID, err)
	}
	if strings.TrimSpace(shop.ID) == "" {
		return shop, ErrAnnouncementShopNotFound
	}
	if !shop.HasLocation() {
		return shop, ErrShopHasNoLocation
	}
	return shop, nil
}

// announcementImageURL picks the picture for the push and makes it absolute.
// An explicit override wins; otherwise the shop's own storefront photo is used,
// which is the whole point of a launch announcement.
func (uc *notificationUseCase) announcementImageURL(shop domain.ShopAnnouncementTarget, override string) string {
	stored := strings.TrimSpace(override)
	if stored == "" {
		stored = strings.TrimSpace(shop.ShopImageURL)
	}
	if stored == "" {
		return ""
	}
	// FCM drops an image it cannot fetch, so a relative path or bare object key
	// must be absolutised here rather than handed over as-is.
	return notificationSvc.ResolvePushImageURL(uc.cloudService, stored)
}

// splitDevices separates the device rows into the unique customers to record an
// in-app notification for, and the flat token list to push to.
func splitDevices(devices []domain.CustomerDevice) (userIDs []string, tokens []string) {
	seenUser := make(map[string]struct{}, len(devices))
	seenToken := make(map[string]struct{}, len(devices))
	for _, d := range devices {
		uid := strings.TrimSpace(d.UserID)
		if uid != "" {
			if _, ok := seenUser[uid]; !ok {
				seenUser[uid] = struct{}{}
				userIDs = append(userIDs, uid)
			}
		}
		tok := strings.TrimSpace(d.Token)
		if tok != "" {
			if _, ok := seenToken[tok]; !ok {
				seenToken[tok] = struct{}{}
				tokens = append(tokens, tok)
			}
		}
	}
	return userIDs, tokens
}
