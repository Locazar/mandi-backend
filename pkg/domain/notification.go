package domain

import "time"

type Notification struct {
	ID                   uint   `gorm:"primaryKey;autoIncrement"`
	SenderType           string `gorm:"type:varchar(50);not null"`
	ReceiverType         string `gorm:"type:varchar(50);not null"`
	Type                 string `gorm:"type:varchar(100);not null"`
	SenderID             uint   `gorm:"not null"`
	Title                string `gorm:"type:varchar(255);not null"`
	Message              string `gorm:"type:text;not null"`
	Body                 string `gorm:"type:text;not null"`
	IsRead               bool   `gorm:"not null;default:false"`
	ReceiverID           uint   `gorm:"not null"`
	CategoryID           uint   `gorm:"not null"`
	ProductID            uint   `gorm:"not null"`
	VariationID          uint   `gorm:"not null"`
	ShopID               uint   `gorm:"not null"`
	UserID               uint   `gorm:"not null"`
	AdminID              uint   `gorm:"not null"`
	OrderID              uint   `gorm:"not null"`
	OfferID              uint   `gorm:"not null"`
	NotificationMetaData string `gorm:"type:text"`
	Status               string `gorm:"type:varchar(50);not null"`
	CreatedAt            string `gorm:"type:varchar(50);not null"`
	UpdatedAt            string `gorm:"type:varchar(50);not null"`
}

type NotificationDeviceToken struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	OwnerID   string     `gorm:"type:varchar(100);not null"`
	OwnerType string     `gorm:"type:varchar(10);not null;check:owner_type IN ('user','seller')"`
	Token     string     `gorm:"type:varchar(255);unique;not null"`
	Platform  string     `gorm:"type:varchar(50)"`
	IsActive  bool       `gorm:"default:true"`
	CreatedAt time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
}
