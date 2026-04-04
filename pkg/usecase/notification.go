package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	notificationSvc "github.com/rohit221990/mandi-backend/pkg/service/notification"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"gorm.io/gorm"
)

type notificationUseCase struct {
	notificationRepo interfaces.NotificationRepository
	fcmPush          notificationSvc.PushSender
	db               *gorm.DB // optional; used to fetch product images for enquiry notifications
}

// NewNotificationUseCase wires a new notification use-case with a lazily-initialised
// FCM push service.  No extra DI provider is required in wire.go.
func NewNotificationUseCase(repo interfaces.NotificationRepository) service.NotificationUseCase {
	return &notificationUseCase{
		notificationRepo: repo,
		fcmPush:          notificationSvc.NewFCMPushService(),
	}
}

// NewNotificationUseCaseWithDB is like NewNotificationUseCase but also accepts a
// GORM database connection.  When provided, enquiry notifications include the
// product item image URLs fetched directly from the SQL database.
func NewNotificationUseCaseWithDB(repo interfaces.NotificationRepository, db *gorm.DB) service.NotificationUseCase {
	return &notificationUseCase{
		notificationRepo: repo,
		fcmPush:          notificationSvc.NewFCMPushService(),
		db:               db,
	}
}

// fetchProductImages queries product_item_images for the given product item ID.
// Returns an empty slice (not an error) when not found.
func (uc *notificationUseCase) fetchProductImages(ctx context.Context, productItemID uint) []string {
	if uc.db == nil || productItemID == 0 {
		return nil
	}
	var images []string
	if err := uc.db.WithContext(ctx).
		Raw(`SELECT product_item_images FROM product_items WHERE id = $1`, productItemID).
		Scan(&images).Error; err != nil {
		log.Printf("WARN [notification]: fetchProductImages id=%d: %v", productItemID, err)
	}
	return images
}

// enquiryDataEnricher returns a DataEnricher func that picks productId / userId
// from the Firestore enquiry document, fetches the product image list from
// PostgreSQL, and adds them to the FCM notification data payload.
func (uc *notificationUseCase) enquiryDataEnricher() func(ctx context.Context, docData map[string]interface{}) map[string]string {
	return func(ctx context.Context, docData map[string]interface{}) map[string]string {
		extra := map[string]string{}

		// Resolve productId — documents may use "productId" or "productItemId".
		productItemID := resolveUintField(docData, "productId", "productItemId", "product_id", "product_item_id")
		if productItemID > 0 {
			extra["product_id"] = strconv.FormatUint(uint64(productItemID), 10)

			images := uc.fetchProductImages(ctx, productItemID)
			if len(images) > 0 {
				extra["product_image_url"] = images[0] // primary image
				if b, err := json.Marshal(images); err == nil {
					extra["product_images"] = string(b) // full list as JSON array
				}
			}
		}

		// Also surface userId so receivers know who initiated.
		if uid := resolveUintField(docData, "userId", "user_id", "customerId"); uid > 0 {
			extra["user_id"] = strconv.FormatUint(uint64(uid), 10)
		}
		if uid := resolveUintField(docData, "sellerId", "shop_id", "sellerId", "seller_id", "customerId"); uid > 0 {
			extra["user_id"] = strconv.FormatUint(uint64(uid), 10)
		}

		return extra
	}
}

// resolveUintField tries each field name in order and returns the first numeric
// value it finds in docData, converted to uint.
func resolveUintField(docData map[string]interface{}, fields ...string) uint {
	for _, f := range fields {
		v, ok := docData[f]
		if !ok {
			continue
		}
		switch n := v.(type) {
		case int64:
			if n > 0 {
				return uint(n)
			}
		case float64:
			if n > 0 {
				return uint(n)
			}
		case int:
			if n > 0 {
				return uint(n)
			}
		case string:
			if u, err := strconv.ParseUint(strings.TrimSpace(n), 10, 64); err == nil && u > 0 {
				return uint(u)
			}
		}
	}
	return 0
}

// SaveNotification persists a notification record to the database.
func (uc *notificationUseCase) SaveNotification(ctx context.Context, n request.Notification) error {
	now := time.Now().UTC().Format(time.RFC3339)
	record := domain.Notification{
		SenderType:           n.SenderType,
		ReceiverType:         n.ReceiverType,
		SenderID:             n.SenderID,
		Title:                n.Title,
		Message:              n.Message,
		Body:                 n.Body,
		IsRead:               false,
		ReceiverID:           n.ReceiverID,
		ShopID:               n.ShopID,
		OrderID:              n.OrderID,
		ProductID:            n.ProductID,
		OfferID:              n.OfferID,
		CategoryID:           n.CategoryID,
		NotificationMetaData: n.NotificationMetaData,
		Status:               n.Status,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := uc.notificationRepo.SaveNotification(ctx, record); err != nil {
		return fmt.Errorf("save notification: %w", err)
	}
	return nil
}

// GetNotificationsBy returns paginated notifications matching the filter.
func (uc *notificationUseCase) GetNotificationsBy(ctx context.Context, filter request.GetNotification, pagination request.Pagination) ([]domain.Notification, error) {
	notifications, err := uc.notificationRepo.GetNotifications(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("get notifications: %w", err)
	}
	return notifications, nil
}

// MarkNotificationAsRead marks a single notification as read.
func (uc *notificationUseCase) MarkNotificationAsRead(ctx context.Context, notificationID uint) error {
	if err := uc.notificationRepo.MarkNotificationAsRead(ctx, notificationID); err != nil {
		return fmt.Errorf("mark notification as read: %w", err)
	}
	return nil
}

// RegisterDeviceToken saves the FCM device token to Postgres and syncs it to
// Firestore so that Cloud Functions can also deliver notifications.
func (uc *notificationUseCase) RegisterDeviceToken(ctx context.Context, req request.NotificationDeviceToken) error {
	// Persist in Postgres
	token := domain.NotificationDeviceToken{
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Token:     req.Token,
		Platform:  req.Platform,
		IsActive:  true,
	}
	if err := uc.notificationRepo.SaveDeviceToken(ctx, token); err != nil {
		return fmt.Errorf("save device token: %w", err)
	}

	// Sync to Firestore (best-effort; don't fail the request on Firestore error)
	ownerCollection := ownerTypeToCollection(req.OwnerType)
	if err := uc.fcmPush.SaveTokenToFirestore(ctx, ownerCollection, req.OwnerID, req.Token, req.Platform); err != nil {
		// Log but don't surface Firestore errors to the client
		log.Printf("WARN [RegisterDeviceToken]: Firestore token sync failed for %s/%s: %v", ownerCollection, req.OwnerID, err)
	}
	return nil
}

// UnregisterDeviceToken deactivates a device token on logout or token refresh.
func (uc *notificationUseCase) UnregisterDeviceToken(ctx context.Context, req request.UnregisterDeviceToken) error {
	if err := uc.notificationRepo.DeleteDeviceToken(ctx, req.OwnerID, req.OwnerType, req.Token); err != nil {
		return fmt.Errorf("delete device token: %w", err)
	}

	ownerCollection := ownerTypeToCollection(req.OwnerType)
	if err := uc.fcmPush.DeleteTokenFromFirestore(ctx, ownerCollection, req.OwnerID, req.Token); err != nil {
		_ = err
	}
	return nil
}

// SendPushNotification sends an FCM push to all active devices belonging to ownerID.
// It looks up tokens from Postgres first; on failure it falls back to Firestore.
func (uc *notificationUseCase) SendPushNotification(ctx context.Context, req request.SendPushRequest) error {
	data := req.Data
	if data == nil {
		data = map[string]string{}
	}
	if req.EventType != "" {
		data["event_type"] = req.EventType
	}

	// Primary path: tokens from Postgres
	tokens, err := uc.notificationRepo.GetActiveTokensByOwner(ctx, req.OwnerID, req.OwnerType)
	if err == nil && len(tokens) > 0 {
		return uc.fcmPush.SendToTokens(ctx, tokens, req.Title, req.Body, data)
	}

	// Fallback: tokens from Firestore (populated by Cloud Functions or other services)
	ownerCollection := ownerTypeToCollection(req.OwnerType)
	return uc.fcmPush.SendToOwnerViaFirestore(ctx, ownerCollection, req.OwnerID, req.Title, req.Body, data)
}

// SendPushToUserOnOrderUpdate is a convenience helper called by the order usecase
// after an order status change.  It builds the payload and delegates to SendPushNotification.
func SendPushToUserOnOrderUpdate(ctx context.Context, uc service.NotificationUseCase, userID uint, orderID uint, newStatus string) {
	req := request.SendPushRequest{
		OwnerID:   strconv.FormatUint(uint64(userID), 10),
		OwnerType: "user",
		Title:     orderStatusTitle(newStatus),
		Body:      orderStatusBody(newStatus, orderID),
		EventType: "order_status_changed",
		Data: map[string]string{
			"order_id": strconv.FormatUint(uint64(orderID), 10),
			"status":   newStatus,
		},
	}
	// Fire-and-forget; don't block the order flow
	go func() {
		_ = uc.SendPushNotification(context.Background(), req)
	}()
}

// SendPushToSellerOnNewOrder notifies a seller when a new order is placed for their shop.
func SendPushToSellerOnNewOrder(ctx context.Context, uc service.NotificationUseCase, shopOwnerID uint, orderID uint) {
	req := request.SendPushRequest{
		OwnerID:   strconv.FormatUint(uint64(shopOwnerID), 10),
		OwnerType: "seller",
		Title:     "New Order Received!",
		Body:      fmt.Sprintf("Order #%d has been placed. Prepare for dispatch.", orderID),
		EventType: "new_order",
		Data: map[string]string{
			"order_id": strconv.FormatUint(uint64(orderID), 10),
		},
	}
	go func() {
		_ = uc.SendPushNotification(context.Background(), req)
	}()
}

// StartFirestoreWatcher starts background Firestore listeners that send FCM
// push notifications when monitored document fields change.
//
// rules may be nil — in that case the default e-commerce rules are used.
//
// Enquiry handling is controlled by ENQUIRY_NOTIFICATION_HANDLER:
//   "server" (default / unset) — this process watches the enquiry collection.
//                                Use this when no Cloud Function is deployed.
//   "cf"                       — skip the enquiry watcher here; the deployed
//                                Cloud Function (ProcessEnquiryUpdate/Create)
//                                is the sole handler. Prevents double delivery.
//
// The method returns as soon as the watcher goroutines are launched; they run
// until ctx is cancelled.
func (uc *notificationUseCase) StartFirestoreWatcher(ctx context.Context, rules []notificationSvc.WatchRule) error {
	if len(rules) == 0 {
		rules = []notificationSvc.WatchRule{
			notificationSvc.DefaultOrderRule(),
			notificationSvc.DefaultProductRule(),
			notificationSvc.DefaultShopRule(),
		}

		// Include the enquiry watcher unless the operator has delegated enquiry
		// notifications to a Cloud Function (ENQUIRY_NOTIFICATION_HANDLER=cf).
		if !isCloudFunctionEnquiryHandler() {
			enquiryRule := notificationSvc.DefaultEnquiryRule()
			if uc.db != nil {
				enquiryRule.DataEnricher = uc.enquiryDataEnricher()
			}
			rules = append(rules, enquiryRule)
			log.Println("INFO [FirestoreWatcher]: enquiry rule enabled (ENQUIRY_NOTIFICATION_HANDLER=server)")
		} else {
			log.Println("INFO [FirestoreWatcher]: enquiry rule disabled — delegated to Cloud Function (ENQUIRY_NOTIFICATION_HANDLER=cf)")
		}
	}

	watcher := notificationSvc.NewFirestoreWatcher(uc.fcmPush, rules...)
	return watcher.Start(ctx)
}

func isCloudFunctionEnquiryHandler() bool {
	mode := strings.Trim(strings.TrimSpace(os.Getenv("ENQUIRY_NOTIFICATION_HANDLER")), `"'`)
	return strings.EqualFold(mode, "cf")
}

// --- helpers ---

func ownerTypeToCollection(ownerType string) string {
	if ownerType == "seller" {
		return "sellers"
	}
	return "users"
}

func orderStatusTitle(status string) string {
	titles := map[string]string{
		"order placed":     "Order Confirmed!",
		"payment pending":  "Payment Required",
		"order cancelled":  "Order Cancelled",
		"order delivered":  "Order Delivered!",
		"return requested": "Return Requested",
		"return approved":  "Return Approved",
		"return cancelled": "Return Cancelled",
		"order returned":   "Order Returned",
	}
	if t, ok := titles[status]; ok {
		return t
	}
	return "Order Update"
}

func orderStatusBody(status string, orderID uint) string {
	bodies := map[string]string{
		"order placed":     fmt.Sprintf("Your order #%d has been confirmed and is being processed.", orderID),
		"payment pending":  fmt.Sprintf("Complete payment for order #%d to confirm your order.", orderID),
		"order cancelled":  fmt.Sprintf("Your order #%d has been cancelled.", orderID),
		"order delivered":  fmt.Sprintf("Your order #%d has been delivered. Enjoy!", orderID),
		"return requested": fmt.Sprintf("Return request for order #%d is being reviewed.", orderID),
		"return approved":  fmt.Sprintf("Your return for order #%d has been approved.", orderID),
		"return cancelled": fmt.Sprintf("Return for order #%d was declined. Contact support for help.", orderID),
		"order returned":   fmt.Sprintf("Your order #%d has been returned and refund initiated.", orderID),
	}
	if b, ok := bodies[status]; ok {
		return b
	}
	return fmt.Sprintf("Your order #%d status has changed to: %s", orderID, status)
}
