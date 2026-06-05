package domain

import (
	"time"

	"gorm.io/gorm"
)

// ReviewType indicates whether a review is for a shop or a product.
type ReviewType string

const (
	ReviewTypeShop    ReviewType = "SHOP"
	ReviewTypeProduct ReviewType = "PRODUCT"
)

func (rt ReviewType) IsValid() bool {
	switch rt {
	case ReviewTypeShop, ReviewTypeProduct:
		return true
	}
	return false
}

// ReactionType indicates the type of reaction to a review.
type ReactionType string

const (
	ReactionTypeLike        ReactionType = "LIKE"
	ReactionTypeHelpful     ReactionType = "HELPFUL"
	ReactionTypeReportAbuse ReactionType = "REPORT_ABUSE"
)

func (r ReactionType) IsValid() bool {
	switch r {
	case ReactionTypeLike, ReactionTypeHelpful, ReactionTypeReportAbuse:
		return true
	}
	return false
}

// FeedbackType indicates the category of feedback.
type FeedbackType string

const (
	FeedbackTypeShop    FeedbackType = "SHOP"
	FeedbackTypeProduct FeedbackType = "PRODUCT"
	FeedbackTypeApp     FeedbackType = "APP"
)

func (ft FeedbackType) IsValid() bool {
	switch ft {
	case FeedbackTypeShop, FeedbackTypeProduct, FeedbackTypeApp:
		return true
	}
	return false
}

// ReviewStatus indicates the status of a review.
type ReviewStatus string

const (
	ReviewStatusActive    ReviewStatus = "active"
	ReviewStatusDeleted   ReviewStatus = "deleted"
	ReviewStatusModerated ReviewStatus = "moderated"
)

func (rs ReviewStatus) IsValid() bool {
	switch rs {
	case ReviewStatusActive, ReviewStatusDeleted, ReviewStatusModerated:
		return true
	}
	return false
}

// FeedbackStatus indicates the status of a feedback.
type FeedbackStatus string

const (
	FeedbackStatusNew      FeedbackStatus = "new"
	FeedbackStatusResolved FeedbackStatus = "resolved"
	FeedbackStatusArchived FeedbackStatus = "archived"
)

func (fs FeedbackStatus) IsValid() bool {
	switch fs {
	case FeedbackStatusNew, FeedbackStatusResolved, FeedbackStatusArchived:
		return true
	}
	return false
}

type ShopReview struct {
	ID         uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID     string         `gorm:"type:varchar(32);not null;index" json:"shop_id"`
	UserID     string         `gorm:"type:varchar(32);not null;index" json:"user_id"`
	Rating     float32        `gorm:"type:decimal(2,1);not null" json:"rating"`
	ReviewText string         `gorm:"type:text;size:2000" json:"review_text"`
	Status     ReviewStatus   `gorm:"type:review_status;default:'active';not null" json:"status"`
	CreatedAt  time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type ProductReview struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductItemID string         `gorm:"type:varchar(32);not null;index" json:"product_item_id"`
	ShopID        string         `gorm:"type:varchar(32);not null;index" json:"shop_id"`
	UserID        string         `gorm:"type:varchar(32);not null;index" json:"user_id"`
	Rating        float32        `gorm:"type:decimal(2,1);not null" json:"rating"`
	ReviewText    string         `gorm:"type:text;size:2000" json:"review_text"`
	Status        ReviewStatus   `gorm:"type:review_status;default:'active';not null" json:"status"`
	CreatedAt     time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type ReviewImage struct {
	ID         uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	ReviewType ReviewType `gorm:"type:review_type;not null" json:"review_type"`
	ReviewID   uint       `gorm:"not null;index" json:"review_id"`
	ImageURL   string     `gorm:"not null" json:"image_url"`
	CreatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

type ReviewComment struct {
	ID              uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ReviewType      ReviewType     `gorm:"type:review_type;not null" json:"review_type"`
	ReviewID        uint           `gorm:"not null;index" json:"review_id"`
	ParentCommentID *uint          `gorm:"index" json:"parent_comment_id"`
	UserID          string         `gorm:"type:varchar(32);not null;index" json:"user_id"`
	Comment         string         `gorm:"type:text;size:1000;not null" json:"comment"`
	CreatedAt       time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type ReviewReaction struct {
	ID           uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	ReviewType   ReviewType   `gorm:"type:review_type;not null" json:"review_type"`
	ReviewID     uint         `gorm:"not null;index" json:"review_id"`
	UserID       string       `gorm:"type:varchar(32);not null;index" json:"user_id"`
	ReactionType ReactionType `gorm:"type:reaction_type;not null" json:"reaction_type"`
	CreatedAt    time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

type Feedback struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	FeedbackType  FeedbackType   `gorm:"type:feedback_type;not null" json:"feedback_type"`
	ShopID        *string        `gorm:"type:varchar(32)" json:"shop_id"`
	ProductItemID *string        `gorm:"type:varchar(32)" json:"product_item_id"`
	UserID        *string        `gorm:"type:varchar(32)" json:"user_id"`
	Rating        *float32       `gorm:"type:decimal(2,1)" json:"rating"`
	FeedbackText  string         `gorm:"type:text;not null" json:"feedback_text"`
	Status        FeedbackStatus `gorm:"type:feedback_status;default:'new';not null" json:"status"`
	Anonymous     bool           `gorm:"not null;default:false" json:"anonymous"`
	CreatedAt     time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

type ShopRatingSummary struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID        string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"shop_id"`
	AverageRating float32   `gorm:"type:decimal(2,1);not null" json:"average_rating"`
	TotalReviews  uint      `gorm:"not null" json:"total_reviews"`
	Rating1Count  uint      `gorm:"not null;default:0" json:"rating_1_count"`
	Rating2Count  uint      `gorm:"not null;default:0" json:"rating_2_count"`
	Rating3Count  uint      `gorm:"not null;default:0" json:"rating_3_count"`
	Rating4Count  uint      `gorm:"not null;default:0" json:"rating_4_count"`
	Rating5Count  uint      `gorm:"not null;default:0" json:"rating_5_count"`
	UpdatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type ProductRatingSummary struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductItemID string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"product_item_id"`
	AverageRating float32   `gorm:"type:decimal(2,1);not null" json:"average_rating"`
	TotalReviews  uint      `gorm:"not null" json:"total_reviews"`
	Rating1Count  uint      `gorm:"not null;default:0" json:"rating_1_count"`
	Rating2Count  uint      `gorm:"not null;default:0" json:"rating_2_count"`
	Rating3Count  uint      `gorm:"not null;default:0" json:"rating_3_count"`
	Rating4Count  uint      `gorm:"not null;default:0" json:"rating_4_count"`
	Rating5Count  uint      `gorm:"not null;default:0" json:"rating_5_count"`
	UpdatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
