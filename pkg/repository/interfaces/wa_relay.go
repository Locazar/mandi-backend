package interfaces

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type WAChatRepository interface {
	EnsureTables(ctx context.Context) error

	// FindShopByPhone looks up a shop whose registered phone matches — used to
	// identify seller inbound messages. Returns gorm.ErrRecordNotFound when no match.
	FindShopByPhone(ctx context.Context, phone string) (domain.ShopDetails, error)

	// FindActiveChat returns the active chat for a given (customerPhone, shopID) pair.
	FindActiveChat(ctx context.Context, customerPhone, shopID string) (domain.WAChat, error)

	// FindActiveChatByCustomerPhone returns the most recent active chat for a customer
	// (used when subsequent messages don't carry a SHOP: token).
	FindActiveChatByCustomerPhone(ctx context.Context, customerPhone string) (domain.WAChat, error)

	// FindActiveChatBySellerPhone returns the most recently active chat where the
	// seller's phone matches — used to route a seller reply back to the right customer.
	FindActiveChatBySellerPhone(ctx context.Context, sellerPhone string) (domain.WAChat, error)

	// FindChatByOutboundMsgID returns the chat whose relayed message has the
	// given Meta message ID — used to route a swipe-reply (context.id).
	FindChatByOutboundMsgID(ctx context.Context, outboundWaMsgID string) (domain.WAChat, error)

	// CountRecentChatsForSeller counts the seller's active chats in which the
	// customer wrote after `since` — used to decide whether an un-quoted
	// seller reply is ambiguous.
	CountRecentChatsForSeller(ctx context.Context, sellerPhone string, since time.Time) (int64, error)

	CreateChat(ctx context.Context, chat domain.WAChat) error
	GetChatByID(ctx context.Context, id string) (domain.WAChat, error)
	ListChats(ctx context.Context, status string, limit, offset int) ([]domain.WAChat, error)

	SaveMessage(ctx context.Context, msg domain.WAMessage) error
	GetMessages(ctx context.Context, chatID string) ([]domain.WAMessage, error)

	MarkAdminNotified(ctx context.Context, chatID string) error
	UpdateLastMessage(ctx context.Context, chatID string, direction string) error
}
