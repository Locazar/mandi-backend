package domain

import "time"

type Notification struct {
	ID                   string `gorm:"primaryKey;type:varchar(32)"`
	SenderType           UserType `gorm:"type:varchar(50);not null"`
	ReceiverType         UserType `gorm:"type:varchar(50);not null"`
	Type                 string `gorm:"type:varchar(100);not null"`
	SenderID             string `gorm:"type:varchar(32);not null"`
	Title                string `gorm:"type:varchar(255);not null"`
	Message              string `gorm:"type:text;not null"`
	Body                 string `gorm:"type:text;not null"`
	IsRead               bool   `gorm:"not null;default:false"`
	ReceiverID           string `gorm:"type:varchar(32);not null"`
	CategoryID           string `gorm:"type:varchar(32);not null"`
	ProductID            string `gorm:"type:varchar(32);not null"`
	VariationID          string `gorm:"type:varchar(32);not null"`
	ShopID               string `gorm:"type:varchar(32);not null"`
	UserID               string `gorm:"type:varchar(32);not null"`
	AdminID              string `gorm:"type:varchar(32);not null"`
	OrderID              string `gorm:"type:varchar(32);not null"`
	OfferID              string `gorm:"type:varchar(32);not null"`
	NotificationMetaData string `gorm:"type:text"`
	Status               NotificationStatus `gorm:"type:varchar(50);not null"`
	CreatedAt            string `gorm:"type:varchar(50);not null"`
	UpdatedAt            string `gorm:"type:varchar(50);not null"`
}

type NotificationDeviceToken struct {
	ID        string     `gorm:"primaryKey;type:varchar(32)"`
	OwnerID   string     `gorm:"type:varchar(100);not null"`
	ShopID    string     `gorm:"type:varchar(100);not null"`
	AdminID   string     `gorm:"type:varchar(100);not null"`
	OwnerType string     `gorm:"type:varchar(10);not null;check:owner_type IN ('user','seller')"`
	Token     string     `gorm:"type:varchar(255);unique;not null"`
	Platform  string     `gorm:"type:varchar(50)"`
	IsActive  bool       `gorm:"default:true"`
	CreatedAt time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
}
