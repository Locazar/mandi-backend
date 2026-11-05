package domain

import "time"

const (
	WAChatStatusActive = "active"
	WAChatStatusClosed = "closed"

	WADirCustomerToSeller = "customer_to_seller"
	WADirSellerToCustomer = "seller_to_customer"
)

// WAChat tracks a relay conversation between a customer and a seller via
// Locazar's WhatsApp Business number. Each unique (customer_phone, shop_id)
// pair gets one active chat at a time.
type WAChat struct {
	ID               string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	CustomerPhone    string     `json:"customer_phone" gorm:"type:varchar(20);not null;index"`
	ShopID           string     `json:"shop_id" gorm:"type:varchar(32);not null;index"`
	ShopName         string     `json:"shop_name" gorm:"type:varchar(255)"`
	SellerPhone      string     `json:"seller_phone" gorm:"type:varchar(20)"`
	Status           string     `json:"status" gorm:"type:varchar(20);default:'active'"`
	AdminNotified    bool       `json:"admin_notified" gorm:"default:false"`
	LastCustomerMsgAt *time.Time `json:"last_customer_msg_at,omitempty"`
	LastSellerMsgAt  *time.Time `json:"last_seller_msg_at,omitempty"`
	LastMessageAt    *time.Time `json:"last_message_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (WAChat) TableName() string { return "wa_chats" }

// WAMessage is one relayed message in a WAChat.
type WAMessage struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ChatID    string    `json:"chat_id" gorm:"type:varchar(36);not null;index"`
	WaMsgID   string    `json:"wa_msg_id" gorm:"type:varchar(255)"` // Meta's message ID
	Direction string    `json:"direction" gorm:"type:varchar(30)"`  // WADirCustomerToSeller | WADirSellerToCustomer
	Body      string    `json:"body" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}

func (WAMessage) TableName() string { return "wa_messages" }

// WAWebhookPayload is the top-level envelope Meta sends to our webhook URL.
type WAWebhookPayload struct {
	Object string         `json:"object"`
	Entry  []WAWebhookEntry `json:"entry"`
}

type WAWebhookEntry struct {
	Changes []WAWebhookChange `json:"changes"`
}

type WAWebhookChange struct {
	Value WAWebhookValue `json:"value"`
	Field string         `json:"field"`
}

type WAWebhookValue struct {
	MessagingProduct string         `json:"messaging_product"`
	Contacts         []WAWHContact  `json:"contacts"`
	Messages         []WAWHMessage  `json:"messages"`
}

type WAWHContact struct {
	Profile WAWHProfile `json:"profile"`
	WaID    string      `json:"wa_id"`
}

type WAWHProfile struct {
	Name string `json:"name"`
}

type WAWHMessage struct {
	From string    `json:"from"` // sender's phone (E.164 without +)
	ID   string    `json:"id"`
	Text *WAWHText `json:"text,omitempty"`
	Type string    `json:"type"` // "text", "image", etc.
}

type WAWHText struct {
	Body string `json:"body"`
}
