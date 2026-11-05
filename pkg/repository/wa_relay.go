package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type waChatRepository struct {
	db *gorm.DB
}

func NewWAChatRepository(db *gorm.DB) interfaces.WAChatRepository {
	return &waChatRepository{db: db}
}

func (r *waChatRepository) EnsureTables(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(&domain.WAChat{}, &domain.WAMessage{})
}

func (r *waChatRepository) FindShopByPhone(ctx context.Context, phone string) (domain.ShopDetails, error) {
	var shop domain.ShopDetails
	err := r.db.WithContext(ctx).
		Table("shop_details").
		Where("phone = ?", phone).
		First(&shop).Error
	return shop, err
}

func (r *waChatRepository) FindActiveChatByCustomerPhone(ctx context.Context, customerPhone string) (domain.WAChat, error) {
	var chat domain.WAChat
	err := r.db.WithContext(ctx).
		Where("customer_phone = ? AND status = ?", customerPhone, domain.WAChatStatusActive).
		Order("last_message_at DESC, created_at DESC").
		First(&chat).Error
	return chat, err
}

func (r *waChatRepository) FindActiveChat(ctx context.Context, customerPhone, shopID string) (domain.WAChat, error) {
	var chat domain.WAChat
	err := r.db.WithContext(ctx).
		Where("customer_phone = ? AND shop_id = ? AND status = ?", customerPhone, shopID, domain.WAChatStatusActive).
		Order("created_at DESC").
		First(&chat).Error
	return chat, err
}

func (r *waChatRepository) FindActiveChatBySellerPhone(ctx context.Context, sellerPhone string) (domain.WAChat, error) {
	var chat domain.WAChat
	err := r.db.WithContext(ctx).
		Where("seller_phone = ? AND status = ?", sellerPhone, domain.WAChatStatusActive).
		Order("last_message_at DESC").
		First(&chat).Error
	return chat, err
}

func (r *waChatRepository) CreateChat(ctx context.Context, chat domain.WAChat) error {
	return r.db.WithContext(ctx).Create(&chat).Error
}

func (r *waChatRepository) GetChatByID(ctx context.Context, id string) (domain.WAChat, error) {
	var chat domain.WAChat
	err := r.db.WithContext(ctx).First(&chat, "id = ?", id).Error
	return chat, err
}

func (r *waChatRepository) ListChats(ctx context.Context, status string, limit, offset int) ([]domain.WAChat, error) {
	q := r.db.WithContext(ctx).Order("last_message_at DESC, created_at DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var chats []domain.WAChat
	return chats, q.Find(&chats).Error
}

func (r *waChatRepository) SaveMessage(ctx context.Context, msg domain.WAMessage) error {
	return r.db.WithContext(ctx).Create(&msg).Error
}

func (r *waChatRepository) GetMessages(ctx context.Context, chatID string) ([]domain.WAMessage, error) {
	var msgs []domain.WAMessage
	err := r.db.WithContext(ctx).Where("chat_id = ?", chatID).Order("created_at ASC").Find(&msgs).Error
	return msgs, err
}

func (r *waChatRepository) MarkAdminNotified(ctx context.Context, chatID string) error {
	return r.db.WithContext(ctx).Model(&domain.WAChat{}).
		Where("id = ?", chatID).
		Update("admin_notified", true).Error
}

func (r *waChatRepository) UpdateLastMessage(ctx context.Context, chatID string, direction string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_message_at": now,
		"updated_at":      now,
	}
	switch direction {
	case domain.WADirCustomerToSeller:
		updates["last_customer_msg_at"] = now
	case domain.WADirSellerToCustomer:
		updates["last_seller_msg_at"] = now
	default:
		return fmt.Errorf("unknown direction: %s", direction)
	}
	return r.db.WithContext(ctx).Model(&domain.WAChat{}).
		Where("id = ?", chatID).
		Updates(updates).Error
}
