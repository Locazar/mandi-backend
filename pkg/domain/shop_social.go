package domain

import "time"

// ShopSocial aggregates followers, ratings, and reviews for a shop in a single table.
type ShopSocial struct {
	ID      uint `json:"id" gorm:"primaryKey;autoIncrement"`
	ShopID  uint `json:"shop_id" gorm:"index;not null"`
	AdminID uint `json:"admin_id" gorm:"index;not null"` // Follower or reviewer
	// Rating: if nonzero, this row is a rating by the user for the shop
	UserID     uint      `json:"user_id" gorm:"index;not null"` // Follower or reviewer
	IsFollower bool      `json:"is_follower" gorm:"type:boolean;default:false"`
	Rating     uint      `json:"rating" gorm:"type:int;default:0"` // 1-5 stars
	Review     string    `json:"review" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// GORM migration: add to AutoMigrate in db/connection.go
// Usage:
// - Follower: IsFollower=true, Rating=0, Review=""
// - Rating:   Rating>0, IsFollower can be true/false, Review=""
// - Review:   Review!="", IsFollower/Rating as appropriate
