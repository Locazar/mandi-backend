package response

import "time"

type Shop struct {
	ID                     uint      `json:"shop_id"`
	ShopName               string    `json:"shop_name"`
	Email                  string    `json:"email"`
	Phone                  string    `json:"phone"`
	AddressLine1           string    `json:"address_line1"`
	AddressLine2           string    `json:"address_line2"`
	City                   string    `json:"city"`
	State                  string    `json:"state"`
	Country                string    `json:"country"`
	Pincode                string    `json:"pincode"`
	ShopType               string    `json:"shop_type"`
	ShopVerificationStatus string    `json:"shop_verification_status"`
	ShopImageURL           string    `json:"shop_image_url"`
	Latitude               float64   `json:"latitude"`
	Longitude              float64   `json:"longitude"`
	FollowerCount          int64     `json:"follower_count"`
	FollowingCount         int64     `json:"following_count"`
	LikeCount              int64     `json:"like_count"`
	RatingCount            int64     `json:"rating_count"`
	ReviewCount            int64     `json:"review_count"`
	AverageRating          float64   `json:"average_rating"`
	IsFollowing            bool      `json:"is_following"`
	IsLiked                bool      `json:"is_liked"`
	UserRating             uint      `json:"user_rating"`
	UserReview             string    `json:"user_review"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}
