package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoi "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/whatsapp"
	uci "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"gorm.io/gorm"
)

const waPageSize = 50

type waChatUseCase struct {
	repo         repoi.WAChatRepository
	adminRepo    repoi.AdminRepository // GetShopByID for new chat creation
	notification uci.NotificationUseCase
	wa           *whatsapp.Client
}

func NewWAChatUseCase(
	repo repoi.WAChatRepository,
	adminRepo repoi.AdminRepository,
	notification uci.NotificationUseCase,
	wa *whatsapp.Client,
) uci.WAChatUseCase {
	return &waChatUseCase{
		repo:         repo,
		adminRepo:    adminRepo,
		notification: notification,
		wa:           wa,
	}
}

func (uc *waChatUseCase) HandleIncoming(ctx context.Context, payload domain.WAWebhookPayload) error {
	if err := uc.repo.EnsureTables(ctx); err != nil {
		return fmt.Errorf("wa_relay: ensure tables: %w", err)
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				if err := uc.routeMessage(ctx, msg); err != nil {
					log.Printf("[wa_relay] route error from=%s: %v", msg.From, err)
				}
			}
		}
	}
	return nil
}

// routeMessage decides if the sender is a customer or a seller and relays.
func (uc *waChatUseCase) routeMessage(ctx context.Context, msg domain.WAWHMessage) error {
	body := ""
	if msg.Text != nil {
		body = msg.Text.Body
	}
	if body == "" && msg.Type != "text" {
		body = fmt.Sprintf("[%s message]", msg.Type)
	}

	// Check if the sender is a known seller (their phone matches a shop).
	shop, err := uc.repo.FindShopByPhone(ctx, msg.From)
	if err == nil && shop.ID != "" {
		return uc.relaySellerToCustomer(ctx, msg.From, shop, body, msg.ID)
	}

	// Otherwise treat as customer.
	return uc.relayCustomerToSeller(ctx, msg.From, body, msg.ID)
}

// relayCustomerToSeller forwards a customer message to the right seller.
// It parses a "SHOP:{id}" token from the message body to identify the shop on
// first contact; on subsequent messages it reuses the most recent active chat.
func (uc *waChatUseCase) relayCustomerToSeller(ctx context.Context, customerPhone, body, waMsgID string) error {
	shopID := extractShopID(body)
	var chat domain.WAChat
	var err error

	if shopID != "" {
		// New chat initiated from shop page.
		shop, shopErr := uc.adminRepo.GetShopByID(ctx, shopID)
		if shopErr != nil {
			return fmt.Errorf("shop not found %s: %w", shopID, shopErr)
		}

		// Reuse existing active chat for this pair, or create a new one.
		chat, err = uc.repo.FindActiveChat(ctx, customerPhone, shopID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			chat = domain.WAChat{
				ID:            uuid.New().String(),
				CustomerPhone: customerPhone,
				ShopID:        shopID,
				ShopName:      shop.ShopName,
				SellerPhone:   shop.Phone,
				Status:        domain.WAChatStatusActive,
			}
			if createErr := uc.repo.CreateChat(ctx, chat); createErr != nil {
				return fmt.Errorf("create chat: %w", createErr)
			}
			uc.notifyAdmin(chat)
		} else if err != nil {
			return fmt.Errorf("find chat: %w", err)
		}
	} else {
		// No SHOP: token — route to the customer's most recent active chat.
		chat, err = uc.repo.FindActiveChatByCustomerPhone(ctx, customerPhone)
		if err != nil {
			return fmt.Errorf("no active chat for customer %s: %w", customerPhone, err)
		}
	}

	waMsg := domain.WAMessage{
		ID:        uuid.New().String(),
		ChatID:    chat.ID,
		WaMsgID:   waMsgID,
		Direction: domain.WADirCustomerToSeller,
		Body:      body,
		CreatedAt: time.Now(),
	}
	if saveErr := uc.repo.SaveMessage(ctx, waMsg); saveErr != nil {
		log.Printf("[wa_relay] save message: %v", saveErr)
	}
	_ = uc.repo.UpdateLastMessage(ctx, chat.ID, domain.WADirCustomerToSeller)

	forwardBody := fmt.Sprintf("💬 Customer message:\n\n%s\n\n_(Reply to this chat to respond)_", body)
	return uc.wa.SendText(chat.SellerPhone, forwardBody)
}

// relaySellerToCustomer forwards a seller reply to the customer of the most
// recently active chat for that seller.
func (uc *waChatUseCase) relaySellerToCustomer(ctx context.Context, sellerPhone string, shop domain.ShopDetails, body, waMsgID string) error {
	chat, err := uc.repo.FindActiveChatBySellerPhone(ctx, sellerPhone)
	if err != nil {
		return fmt.Errorf("no active chat for seller %s: %w", sellerPhone, err)
	}

	waMsg := domain.WAMessage{
		ID:        uuid.New().String(),
		ChatID:    chat.ID,
		WaMsgID:   waMsgID,
		Direction: domain.WADirSellerToCustomer,
		Body:      body,
		CreatedAt: time.Now(),
	}
	if saveErr := uc.repo.SaveMessage(ctx, waMsg); saveErr != nil {
		log.Printf("[wa_relay] save message: %v", saveErr)
	}
	_ = uc.repo.UpdateLastMessage(ctx, chat.ID, domain.WADirSellerToCustomer)

	forwardBody := fmt.Sprintf("Reply from *%s*:\n\n%s", shop.ShopName, body)
	return uc.wa.SendText(chat.CustomerPhone, forwardBody)
}

func (uc *waChatUseCase) notifyAdmin(chat domain.WAChat) {
	req := request.SendPushRequest{
		OwnerID:   "admin",
		OwnerType: "admin",
		Title:     "💬 New WhatsApp chat",
		Body:      fmt.Sprintf("A customer started a WhatsApp chat with %s", chat.ShopName),
		EventType: "wa_chat_started",
		Data: map[string]string{
			"chat_id": chat.ID,
			"shop_id": chat.ShopID,
		},
	}
	if _, err := uc.notification.SendPushNotification(context.Background(), req); err != nil {
		log.Printf("[wa_relay] admin notify: %v", err)
	}
	_ = uc.repo.MarkAdminNotified(context.Background(), chat.ID)
}

func (uc *waChatUseCase) ListChats(ctx context.Context, status string, page int) ([]domain.WAChat, error) {
	if err := uc.repo.EnsureTables(ctx); err != nil {
		return nil, err
	}
	offset := (page - 1) * waPageSize
	if offset < 0 {
		offset = 0
	}
	return uc.repo.ListChats(ctx, status, waPageSize, offset)
}

func (uc *waChatUseCase) GetChat(ctx context.Context, chatID string) (domain.WAChat, []domain.WAMessage, error) {
	chat, err := uc.repo.GetChatByID(ctx, chatID)
	if err != nil {
		return chat, nil, err
	}
	msgs, err := uc.repo.GetMessages(ctx, chatID)
	return chat, msgs, err
}

// extractShopID looks for "SHOP:{id}" anywhere in a message body.
func extractShopID(body string) string {
	upper := strings.ToUpper(body)
	const prefix = "SHOP:"
	idx := strings.Index(upper, prefix)
	if idx == -1 {
		return ""
	}
	rest := body[idx+len(prefix):]
	end := strings.IndexAny(rest, " \n\t\r")
	if end == -1 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:end])
}
