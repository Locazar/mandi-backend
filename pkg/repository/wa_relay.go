package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type waChatRepository struct {
	db *gorm.DB

	ensureMu sync.Mutex
	ensured  bool
}

func NewWAChatRepository(db *gorm.DB) interfaces.WAChatRepository {
	return &waChatRepository{db: db}
}

// waRelaySchema creates or upgrades the relay tables. Plain idempotent SQL is
// used instead of GORM AutoMigrate: AutoMigrate's column introspection
// ("SELECT * FROM wa_chats LIMIT 1") fails on this database with
// "insufficient arguments", which stopped every inbound message.
var waRelaySchema = []string{
	`CREATE TABLE IF NOT EXISTS wa_chats (
		id                   varchar(36) PRIMARY KEY,
		customer_phone       varchar(20) NOT NULL,
		shop_id              varchar(32) NOT NULL,
		shop_name            varchar(255),
		seller_phone         varchar(20),
		status               varchar(20) DEFAULT 'active',
		admin_notified       boolean DEFAULT false,
		last_customer_msg_at timestamptz,
		last_seller_msg_at   timestamptz,
		last_message_at      timestamptz,
		created_at           timestamptz,
		updated_at           timestamptz
	)`,
	`CREATE TABLE IF NOT EXISTS wa_messages (
		id         varchar(36) PRIMARY KEY,
		chat_id    varchar(36) NOT NULL,
		wa_msg_id  varchar(255),
		direction  varchar(30),
		body       text,
		created_at timestamptz
	)`,
	`ALTER TABLE wa_messages ADD COLUMN IF NOT EXISTS outbound_wa_msg_id varchar(255)`,
	`CREATE INDEX IF NOT EXISTS idx_wa_chats_customer_phone ON wa_chats (customer_phone)`,
	`CREATE INDEX IF NOT EXISTS idx_wa_chats_shop_id ON wa_chats (shop_id)`,
	`CREATE INDEX IF NOT EXISTS idx_wa_chats_seller_phone ON wa_chats (seller_phone)`,
	`CREATE INDEX IF NOT EXISTS idx_wa_messages_chat_id ON wa_messages (chat_id)`,
	`CREATE INDEX IF NOT EXISTS idx_wa_messages_outbound_wa_msg_id ON wa_messages (outbound_wa_msg_id)`,
}

// EnsureTables runs the schema once per process; later calls are free.
func (r *waChatRepository) EnsureTables(ctx context.Context) error {
	r.ensureMu.Lock()
	defer r.ensureMu.Unlock()
	if r.ensured {
		return nil
	}
	for _, stmt := range waRelaySchema {
		if err := r.db.WithContext(ctx).Exec(stmt).Error; err != nil {
			return fmt.Errorf("wa_relay schema: %w", err)
		}
	}
	r.ensured = true
	return nil
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

func (r *waChatRepository) FindChatByOutboundMsgID(ctx context.Context, outboundWaMsgID string) (domain.WAChat, error) {
	var chat domain.WAChat
	err := r.db.WithContext(ctx).
		Table("wa_chats").
		Select("wa_chats.*").
		Joins("JOIN wa_messages m ON m.chat_id = wa_chats.id").
		Where("m.outbound_wa_msg_id = ?", outboundWaMsgID).
		Order("m.created_at DESC").
		Limit(1).
		Scan(&chat).Error
	if err == nil && chat.ID == "" {
		err = gorm.ErrRecordNotFound
	}
	return chat, err
}

func (r *waChatRepository) CountRecentChatsForSeller(ctx context.Context, sellerPhone string, since time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&domain.WAChat{}).
		Where("seller_phone = ? AND status = ? AND last_customer_msg_at > ?",
			sellerPhone, domain.WAChatStatusActive, since).
		Count(&n).Error
	return n, err
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
