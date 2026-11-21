package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoi "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/whatsapp"
	uci "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const waPageSize = 50

// Customer-facing and seller-facing copy for the WhatsApp relay.
const (
	waMsgCustomerAck = "Thanks for your enquiry! The seller will respond in a few minutes. – Team Locazar"

	waMsgCustomerNoChat = "Hi! To message a shop, open it on Locazar and tap *Chat on WhatsApp*. " +
		"Your number stays private and the shop's number stays private too. – Team Locazar"

	waMsgSellerAmbiguous = "You have open enquiries from more than one customer. " +
		"Please *swipe right on the customer's message* and reply, so we know who to send your answer to."

	waMsgSellerNoChat = "There is no open customer enquiry to reply to right now. " +
		"New enquiries will appear here as they arrive."

	waMsgSellerCustomerGone = "We couldn't deliver your reply: the customer hasn't messaged in the last 24 hours, " +
		"so WhatsApp won't let us send it. Please follow up in the Locazar seller app."
)

// waAmbiguityWindow is how far back a customer message counts as "open" when
// deciding whether an un-quoted seller reply is ambiguous.
const waAmbiguityWindow = 24 * time.Hour

var (
	// 8+ digits, optionally separated by spaces, dashes, dots or brackets,
	// with an optional leading + — covers 98765 43210, +91-98765-43210 etc.
	waPhonePattern = regexp.MustCompile(`\+?\d(?:[\s\-.()]*\d){7,}`)
	waEmailPattern = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	waLinkPattern  = regexp.MustCompile(`(?i)\b(?:https?://)?(?:wa\.me|api\.whatsapp\.com|chat\.whatsapp\.com)/\S*`)
)

type waChatUseCase struct {
	repo         repoi.WAChatRepository
	adminRepo    repoi.AdminRepository // GetShopByID for new chat creation
	notification uci.NotificationUseCase
	wa           *whatsapp.Client

	// sellerTemplate is an approved WhatsApp template used to reach a seller
	// who has not messaged the Locazar number in the last 24 hours (Meta
	// rejects free-form messages then). Body: {{1}} = shop name,
	// {{2}} = customer reference, {{3}} = customer message.
	sellerTemplate     string
	sellerTemplateLang string

	// maskContacts hides phone numbers, emails and WhatsApp links typed in
	// relayed messages. Off by default: customers and sellers may share
	// numbers. Set WA_RELAY_MASK_CONTACTS=true to switch masking back on.
	maskContacts bool
}

func NewWAChatUseCase(
	repo repoi.WAChatRepository,
	adminRepo repoi.AdminRepository,
	notification uci.NotificationUseCase,
	wa *whatsapp.Client,
) uci.WAChatUseCase {
	lang := waSetting("WA_SELLER_ENQUIRY_TEMPLATE_LANG")
	if lang == "" {
		lang = "en"
	}
	return &waChatUseCase{
		repo:               repo,
		adminRepo:          adminRepo,
		notification:       notification,
		wa:                 wa,
		sellerTemplate:     waSetting("WA_SELLER_ENQUIRY_TEMPLATE"),
		sellerTemplateLang: lang,
		maskContacts:       strings.EqualFold(waSetting("WA_RELAY_MASK_CONTACTS"), "true"),
	}
}

// waSetting reads a relay setting from the .env file (loaded by viper in
// config.LoadConfig) or, failing that, from the process environment.
func waSetting(key string) string {
	if v := strings.TrimSpace(viper.GetString(key)); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(key))
}

func (uc *waChatUseCase) HandleIncoming(ctx context.Context, payload domain.WAWebhookPayload) error {
	if err := uc.repo.EnsureTables(ctx); err != nil {
		return fmt.Errorf("wa_relay: ensure tables: %w", err)
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			names := make(map[string]string, len(change.Value.Contacts))
			for _, c := range change.Value.Contacts {
				names[c.WaID] = c.Profile.Name
			}
			for _, msg := range change.Value.Messages {
				if err := uc.routeMessage(ctx, msg, names[msg.From]); err != nil {
					log.Printf("[wa_relay] route error msg=%s: %v", msg.ID, err)
				}
			}
		}
	}
	return nil
}

// routeMessage decides if the sender is a customer or a seller and relays.
// Both parties only ever talk to the Locazar number, so neither side's
// WhatsApp number is revealed automatically. Numbers they choose to type, or
// contact cards they share, are relayed as-is unless WA_RELAY_MASK_CONTACTS=true.
func (uc *waChatUseCase) routeMessage(ctx context.Context, msg domain.WAWHMessage, profileName string) error {
	body := ""
	if msg.Text != nil {
		body = msg.Text.Body
	}
	if body == "" && msg.Type == "contacts" && len(msg.Contacts) > 0 {
		body = contactCardsText(msg.Contacts)
	}
	if body == "" && msg.Type != "text" {
		body = fmt.Sprintf("[%s message — media is not forwarded, please describe it in text]", msg.Type)
	}
	quotedID := ""
	if msg.Context != nil {
		quotedID = msg.Context.ID
	}
	log.Printf("[wa_relay] inbound msg=%s from=…%s type=%s shop=%q quoted=%t",
		msg.ID, last4(msg.From), msg.Type, extractShopID(body), quotedID != "")
	if !uc.wa.Enabled() {
		log.Printf("[wa_relay] %v — cannot reply or relay", whatsapp.ErrNotConfigured)
	}

	// A message carrying a SHOP: token is always a customer enquiry, even if
	// the sender is also a registered seller (sellers shop too).
	if extractShopID(body) == "" {
		// Meta delivers msg.From in E.164 without '+' (e.g. 919876543210).
		// DB stores phones as 10-digit Indian numbers, so strip the country
		// code before the lookup.
		shop, err := uc.repo.FindShopByPhone(ctx, strip91(msg.From))
		if err == nil && shop.ID != "" {
			return uc.relaySellerToCustomer(ctx, msg.From, shop, body, msg.ID, quotedID)
		}
	}
	return uc.relayCustomerToSeller(ctx, msg.From, profileName, body, msg.ID, quotedID)
}

// relayCustomerToSeller forwards a customer message to the right seller.
// Routing order: (1) the seller message the customer swipe-replied to,
// (2) a SHOP:{id} token from the shop page, (3) the customer's most recent
// active chat.
func (uc *waChatUseCase) relayCustomerToSeller(ctx context.Context, customerPhone, customerName, body, waMsgID, quotedID string) error {
	var (
		chat    domain.WAChat
		err     error
		found   bool
		isFirst bool
	)

	if quotedID != "" {
		if c, qErr := uc.repo.FindChatByOutboundMsgID(ctx, quotedID); qErr == nil && c.CustomerPhone == customerPhone {
			chat, found = c, true
		}
	}

	if !found {
		if shopID := extractShopID(body); shopID != "" {
			shop, shopErr := uc.adminRepo.GetShopByID(ctx, shopID)
			if shopErr != nil {
				uc.sendSafe(customerPhone, waMsgCustomerNoChat)
				return fmt.Errorf("shop not found %s: %w", shopID, shopErr)
			}
			chat, err = uc.repo.FindActiveChat(ctx, customerPhone, shopID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				chat = domain.WAChat{
					ID:            uuid.New().String(),
					CustomerPhone: customerPhone,
					ShopID:        shopID,
					ShopName:      shop.ShopName,
					// E.164 so FindActiveChatBySellerPhone matches msg.From.
					SellerPhone: toE164IN(shop.Phone),
					Status:      domain.WAChatStatusActive,
				}
				if createErr := uc.repo.CreateChat(ctx, chat); createErr != nil {
					return fmt.Errorf("create chat: %w", createErr)
				}
				isFirst = true
				uc.notifyAdmin(chat)
			} else if err != nil {
				return fmt.Errorf("find chat: %w", err)
			}
			found = true
		}
	}

	if !found {
		chat, err = uc.repo.FindActiveChatByCustomerPhone(ctx, customerPhone)
		if err != nil {
			uc.sendSafe(customerPhone, waMsgCustomerNoChat)
			return fmt.Errorf("no active chat for customer: %w", err)
		}
	}

	// Acknowledge the customer when a chat starts, and again whenever they
	// come back after a quiet day, so they always know the message landed.
	sendAck := isFirst || chat.LastCustomerMsgAt == nil ||
		time.Since(*chat.LastCustomerMsgAt) > waAmbiguityWindow

	cleanBody := uc.relayText(stripShopToken(body))
	ref := chatRef(chat.ID)
	who := "Customer " + ref
	if n := strings.TrimSpace(customerName); n != "" {
		who = fmt.Sprintf("%s (%s)", n, ref)
	}

	var forwardBody string
	if isFirst {
		forwardBody = fmt.Sprintf("💬 *New enquiry for %s via Locazar*\nFrom: %s\n\n%s\n\n↩️ To reply, swipe right on this message and type your answer.",
			chat.ShopName, who, cleanBody)
	} else {
		forwardBody = fmt.Sprintf("💬 *%s:*\n\n%s\n\n↩️ Swipe right on this message to reply.", who, cleanBody)
	}

	outID, sendErr := uc.wa.SendTextWithID(chat.SellerPhone, forwardBody)
	if whatsapp.IsOutsideWindow(sendErr) && uc.sellerTemplate != "" {
		// The seller hasn't messaged the Locazar number in 24h, so only a
		// template can reach them. Swipe-replies on it still route back.
		outID, sendErr = uc.wa.SendTemplate(chat.SellerPhone, uc.sellerTemplate, uc.sellerTemplateLang,
			chat.ShopName, who, cleanBody)
	}

	uc.saveMessage(ctx, chat.ID, waMsgID, outID, domain.WADirCustomerToSeller, body)

	if sendErr != nil {
		log.Printf("[wa_relay] forward to seller failed chat=%s shop=%s: %v", chat.ID, chat.ShopID, sendErr)
		if whatsapp.IsOutsideWindow(sendErr) && uc.sellerTemplate == "" {
			log.Printf("[wa_relay] seller is outside the 24h window and WA_SELLER_ENQUIRY_TEMPLATE is not set")
		}
		uc.notifyAdminUndelivered(chat)
	} else {
		log.Printf("[wa_relay] relayed customer→seller chat=%s shop=%s", chat.ID, chat.ShopID)
	}
	if sendAck {
		uc.sendSafe(customerPhone, waMsgCustomerAck)
	}
	return sendErr
}

// relaySellerToCustomer forwards a seller reply to the right customer.
// Routing order: (1) the customer message the seller swipe-replied to,
// (2) the seller's only open chat. With several open chats and no quote the
// reply is NOT guessed — the seller is asked to swipe-reply instead.
func (uc *waChatUseCase) relaySellerToCustomer(ctx context.Context, sellerPhone string, shop domain.ShopDetails, body, waMsgID, quotedID string) error {
	var (
		chat  domain.WAChat
		found bool
	)

	if quotedID != "" {
		if c, err := uc.repo.FindChatByOutboundMsgID(ctx, quotedID); err == nil && c.SellerPhone == sellerPhone {
			chat, found = c, true
		}
	}

	if !found {
		open, err := uc.repo.CountRecentChatsForSeller(ctx, sellerPhone, time.Now().Add(-waAmbiguityWindow))
		if err != nil {
			return fmt.Errorf("count seller chats: %w", err)
		}
		if open > 1 {
			uc.sendSafe(sellerPhone, waMsgSellerAmbiguous)
			return nil
		}
		chat, err = uc.repo.FindActiveChatBySellerPhone(ctx, sellerPhone)
		if err != nil {
			uc.sendSafe(sellerPhone, waMsgSellerNoChat)
			return fmt.Errorf("no active chat for seller: %w", err)
		}
	}

	shopName := chat.ShopName
	if shopName == "" {
		shopName = shop.ShopName
	}
	forwardBody := fmt.Sprintf("🏪 *%s* replied:\n\n%s\n\n↩️ Swipe right on this message to reply.",
		shopName, uc.relayText(body))

	outID, sendErr := uc.wa.SendTextWithID(chat.CustomerPhone, forwardBody)
	uc.saveMessage(ctx, chat.ID, waMsgID, outID, domain.WADirSellerToCustomer, body)

	if sendErr == nil {
		log.Printf("[wa_relay] relayed seller→customer chat=%s", chat.ID)
	}
	if sendErr != nil {
		log.Printf("[wa_relay] forward to customer failed chat=%s: %v", chat.ID, sendErr)
		if whatsapp.IsOutsideWindow(sendErr) {
			uc.sendSafe(sellerPhone, waMsgSellerCustomerGone)
		}
	}
	return sendErr
}

// saveMessage stores the original (unmasked) text for the admin audit trail
// plus the ID of the relayed copy, then bumps the chat's activity times.
func (uc *waChatUseCase) saveMessage(ctx context.Context, chatID, waMsgID, outboundID, direction, body string) {
	waMsg := domain.WAMessage{
		ID:              uuid.New().String(),
		ChatID:          chatID,
		WaMsgID:         waMsgID,
		OutboundWaMsgID: outboundID,
		Direction:       direction,
		Body:            body,
		CreatedAt:       time.Now(),
	}
	if err := uc.repo.SaveMessage(ctx, waMsg); err != nil {
		log.Printf("[wa_relay] save message: %v", err)
	}
	_ = uc.repo.UpdateLastMessage(ctx, chatID, direction)
}

// sendSafe sends a system message from the bot, logging instead of failing.
func (uc *waChatUseCase) sendSafe(to, text string) {
	if err := uc.wa.SendText(to, text); err != nil {
		log.Printf("[wa_relay] system message failed: %v", err)
	}
}

func (uc *waChatUseCase) notifyAdminUndelivered(chat domain.WAChat) {
	req := request.SendPushRequest{
		OwnerID:   "admin",
		OwnerType: "admin",
		Title:     "⚠️ WhatsApp enquiry not delivered",
		Body:      fmt.Sprintf("A customer enquiry for %s could not be delivered to the seller on WhatsApp", chat.ShopName),
		EventType: "wa_chat_undelivered",
		Data: map[string]string{
			"chat_id": chat.ID,
			"shop_id": chat.ShopID,
		},
	}
	if _, err := uc.notification.SendPushNotification(context.Background(), req); err != nil {
		log.Printf("[wa_relay] admin notify: %v", err)
	}
}

// last4 returns the last four digits of a phone number for logs, so full
// numbers never end up in log files.
func last4(phone string) string {
	if len(phone) <= 4 {
		return phone
	}
	return phone[len(phone)-4:]
}

// chatRef is a short, stable label for a chat (e.g. "#3F2A") so a seller can
// tell customers apart without seeing their numbers.
func chatRef(chatID string) string {
	id := strings.ReplaceAll(chatID, "-", "")
	if len(id) > 4 {
		id = id[:4]
	}
	return "#" + strings.ToUpper(id)
}

// relayText applies contact masking only when WA_RELAY_MASK_CONTACTS=true.
func (uc *waChatUseCase) relayText(s string) string {
	if uc.maskContacts {
		return maskContactDetails(s)
	}
	return s
}

// contactCardsText turns shared contact cards into plain text so they can be
// relayed from the Locazar number, e.g. "📇 Ramesh: +91 98765 43210".
func contactCardsText(cards []domain.WAWHContactCard) string {
	var lines []string
	for _, c := range cards {
		var phones []string
		for _, p := range c.Phones {
			if p.Phone != "" {
				phones = append(phones, p.Phone)
			}
		}
		name := strings.TrimSpace(c.Name.FormattedName)
		if name == "" {
			name = "Contact"
		}
		lines = append(lines, fmt.Sprintf("📇 %s: %s", name, strings.Join(phones, ", ")))
	}
	return strings.Join(lines, "\n")
}

// maskContactDetails hides phone numbers, email addresses and WhatsApp links
// typed inside a relayed message, so neither side can swap direct contacts.
func maskContactDetails(s string) string {
	s = waLinkPattern.ReplaceAllString(s, "[link hidden]")
	s = waEmailPattern.ReplaceAllString(s, "[email hidden]")
	s = waPhonePattern.ReplaceAllString(s, "[number hidden]")
	return s
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

// toE164IN converts an Indian phone stored in the DB (10 digits) to the E.164
// format required by the WhatsApp Cloud API (country code + number, no +).
// Numbers that are already 12 digits starting with 91, or that contain a +
// prefix, pass through cleanly.
func toE164IN(phone string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)
	if len(digits) == 10 {
		return "91" + digits
	}
	return digits
}

// strip91 is the inverse of toE164IN: it strips the Indian country code from a
// 12-digit E.164 number (91XXXXXXXXXX → XXXXXXXXXX) so it matches the 10-digit
// format phones are stored as in the DB.
func strip91(phone string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)
	if len(digits) == 12 && strings.HasPrefix(digits, "91") {
		return digits[2:]
	}
	return digits
}

// stripShopToken removes the "SHOP:{id}" routing token (and its leading blank
// line) from a message body before forwarding it to the seller. The token is
// added by the web app to identify the shop but is an implementation detail
// the seller should never see.
func stripShopToken(body string) string {
	upper := strings.ToUpper(body)
	const prefix = "SHOP:"
	idx := strings.Index(upper, prefix)
	if idx == -1 {
		return body
	}
	// Find the end of the token (next whitespace or end of string).
	rest := body[idx+len(prefix):]
	end := strings.IndexAny(rest, " \n\t\r")
	var after string
	if end == -1 {
		after = ""
	} else {
		after = rest[end:]
	}
	return strings.TrimRight(body[:idx]+after, "\n\r \t")
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
