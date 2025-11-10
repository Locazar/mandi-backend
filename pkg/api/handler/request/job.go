package request

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	JobID       uuid.UUID `json:"job_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CategoryID  uuid.UUID `json:"category_id"`
	LocationID  uuid.UUID `json:"location_id"`
	Company     string    `json:"company"`
	PostedDate  time.Time `json:"posted_date"`
	ExpiryDate  time.Time `json:"expiry_date"`
	IsActive    bool      `json:"is_active"`
}
