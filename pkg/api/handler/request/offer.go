package request

import (
	"time"
)

// CustomTime handles the specific time format from frontend
type CustomTime struct {
	time.Time
}

// UnmarshalJSON handles the time format "2006-01-02T15:04:05.000"
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := string(b[1 : len(b)-1]) // Remove quotes
	t, err := time.Parse("2006-01-02T15:04:05.000", s)
	if err != nil {
		return err
	}
	ct.Time = t
	return nil
}

// offer
type Offer struct {
	Name         string    `json:"offer_name" binding:"required"`
	Description  string    `json:"description" binding:"required,min=6,max=50"`
	DiscountRate uint      `json:"discount_rate" binding:"required,numeric,min=1,max=100"`
	StartDate    time.Time `json:"start_date" binding:"required"`
	EndDate      time.Time `json:"end_date" binding:"required,gtfield=StartDate"`
	Type         string    `json:"offer_type" binding:"required"`
}
type OfferCategory struct {
	OfferID    uint `json:"offer_id" binding:"required"`
	CategoryID uint `json:"category_id" binding:"required"`
}

type OfferProduct struct {
	OfferID       uint `json:"offer_id" binding:"required"`
	ProductItemID uint `json:"product_item_id" binding:"required"`
}

type UpdateCategoryOffer struct {
	CategoryOfferID uint `json:"category_offer_id" binding:"required"`
	OfferID         uint `json:"offer_id" binding:"required"`
}

type UpdateProductOffer struct {
	ProductOfferID uint `json:"product_offer_id" binding:"required"`
	OfferID        uint `json:"offer_id" binding:"required"`
}

type ApplyOfferToShop struct {
	StartDate CustomTime `json:"start_date" binding:"required"`
	EndDate   CustomTime `json:"end_date" binding:"required"`
	OfferID   uint       `json:"offer_id" binding:"required"`
}
