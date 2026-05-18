package request

// ShopRatingRequest is used to create or update a user's rating for a shop.
type ShopRatingRequest struct {
	Rating uint `json:"rating" binding:"required,min=1,max=5"`
}

// ShopReviewRequest is used to create or update a user's review for a shop.
type ShopReviewRequest struct {
	Review string `json:"review" binding:"required"`
}
