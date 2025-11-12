package request

import (
	"github.com/google/uuid"
)

type JobCategory struct {
	CategoryID uuid.UUID  `json:"category_id"`
	Name       string     `json:"name"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
}
