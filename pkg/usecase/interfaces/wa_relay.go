package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type WAChatUseCase interface {
	// HandleIncoming routes an inbound webhook message from Meta.
	// It determines if the sender is a customer or a seller and relays accordingly.
	HandleIncoming(ctx context.Context, payload domain.WAWebhookPayload) error

	// ListChats returns WA chats for the admin portal. status "" returns all.
	ListChats(ctx context.Context, status string, page int) ([]domain.WAChat, error)

	// GetChat returns a single chat with its messages.
	GetChat(ctx context.Context, chatID string) (domain.WAChat, []domain.WAMessage, error)
}
