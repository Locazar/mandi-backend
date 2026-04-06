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
	IsLiked    bool      `json:"is_liked" gorm:"type:boolean;default:false"`
	Rating     uint      `json:"rating" gorm:"type:int;default:0"` // 1-5 stars
	Review     string    `json:"review" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ShopSocialSummary is an aggregated social payload for a shop and a user context.
type ShopSocialSummary struct {
	ShopID         uint    `json:"shop_id"`
	FollowerCount  int64   `json:"follower_count"`
	FollowingCount int64   `json:"following_count"`
	LikeCount      int64   `json:"like_count"`
	RatingCount    int64   `json:"rating_count"`
	ReviewCount    int64   `json:"review_count"`
	AverageRating  float64 `json:"average_rating"`
	IsFollowing    bool    `json:"is_following"`
	IsLiked        bool    `json:"is_liked"`
	UserRating     uint    `json:"user_rating"`
	UserReview     string  `json:"user_review"`
}

// ShopRatingDistribution represents ratings grouped by star value.
type ShopRatingDistribution struct {
	Rating uint  `json:"rating"`
	Count  int64 `json:"count"`
}

// GORM migration: add to AutoMigrate in db/connection.go
// Usage:
// - Follower: IsFollower=true, Rating=0, Review=""
// - Rating:   Rating>0, IsFollower can be true/false, Review=""
// - Review:   Review!="", IsFollower/Rating as appropriate
