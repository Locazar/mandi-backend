package request

// ShopRatingRequest is used to create or update a user's rating for a shop.
type ShopRatingRequest struct {
	Rating uint `json:"rating" binding:"required,min=1,max=5"`
}

// ShopReviewRequest is used to create or update a user's review for a shop.
type ShopReviewRequest struct {
	Review string `json:"review" binding:"required"`
}

// ShopRatingAndReviewRequest is used to create or update both rating and review for a shop in a single request.
// All fields are optional, allowing users to update rating, review, or comments independently.
type ShopRatingAndReviewRequest struct {
	Rating    *uint   `json:"rating" binding:"omitempty,min=1,max=5"`              // Optional: rating from 1-5
	Review    *string `json:"review"`                                              // Optional: review text
	Comments  *string `json:"comments"`                                            // Optional: comments text
	ShopID    *string `json:"shop_id"`                                             // Optional: shop ID (can also come from URL)
	CustomerID *string `json:"customer_id"`                                        // Optional: customer ID
}
