package response

import "time"

type ShopReview struct {
	UserID    uint      `json:"user_id"`
	Rating    uint      `json:"rating"`
	Review    string    `json:"review"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Shop struct {
	ID                     uint      `json:"shop_id" gorm:"column:id"`
	ShopName               string    `json:"shop_name" gorm:"column:shop_name"`
	Email                  string    `json:"email" gorm:"column:email"`
	Phone                  string    `json:"phone" gorm:"column:phone"`
	AddressLine1           string    `json:"address_line1" gorm:"column:address_line1"`
	AddressLine2           string    `json:"address_line2" gorm:"column:address_line2"`
	City                   string    `json:"city" gorm:"column:city"`
	State                  string    `json:"state" gorm:"column:state"`
	Country                string    `json:"country" gorm:"column:country"`
	Pincode                string    `json:"pincode" gorm:"column:pincode"`
	ShopType               string    `json:"shop_type" gorm:"column:shop_type"`
	ShopVerificationStatus string    `json:"shop_verification_status" gorm:"column:shop_verification_status"`
	ShopImageURL           string    `json:"shop_image_url" gorm:"column:shop_image_url"`
	Latitude               float64   `json:"latitude" gorm:"column:latitude"`
	Longitude              float64   `json:"longitude" gorm:"column:longitude"`
	FollowerCount          int64     `json:"follower_count" gorm:"column:follower_count"`
	FollowingCount         int64     `json:"following_count" gorm:"column:following_count"`
	LikeCount              int64     `json:"like_count" gorm:"column:like_count"`
	RatingCount            int64     `json:"rating_count" gorm:"column:rating_count"`
	ReviewCount            int64     `json:"review_count" gorm:"column:review_count"`
	AverageRating          float64   `json:"average_rating" gorm:"column:average_rating"`
	IsFollowing            bool      `json:"is_following" gorm:"column:is_following"`
	IsLiked                bool      `json:"is_liked" gorm:"column:is_liked"`
	UserRating             uint      `json:"user_rating" gorm:"column:user_rating"`
	UserReview             string    `json:"user_review" gorm:"column:user_review"`
	DistanceKm             *float64  `json:"distance_km,omitempty" gorm:"column:distance_km"`
	IsOpen                 bool          `json:"is_open" gorm:"column:is_open"`
	CreatedAt              time.Time     `json:"created_at" gorm:"column:created_at"`
	UpdatedAt              time.Time     `json:"updated_at" gorm:"column:updated_at"`
	Reviews                []ShopReview  `json:"reviews"`
}
