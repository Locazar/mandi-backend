package domain

import "time"

// CategoryRequestStatus is the review state of a seller's request for a new
// department/category.
type CategoryRequestStatus string

const (
	CategoryRequestPending  CategoryRequestStatus = "pending"
	CategoryRequestApproved CategoryRequestStatus = "approved"
	CategoryRequestRejected CategoryRequestStatus = "rejected"
)

func (s CategoryRequestStatus) IsValid() bool {
	switch s {
	case CategoryRequestPending, CategoryRequestApproved, CategoryRequestRejected:
		return true
	}
	return false
}

// CategoryRequest is a seller-submitted ask for a department/category that
// isn't in the existing list — surfaced from the seller-app's "More"
// section, reviewed by a platform user (super_admin/catalog manager) from
// admin-portal.
type CategoryRequest struct {
	ID             string                `json:"id" gorm:"primaryKey;type:varchar(32)"`
	AdminID        string                `json:"admin_id" gorm:"type:varchar(32);not null;index"`
	ShopID         string                `json:"shop_id,omitempty" gorm:"type:varchar(32);index"`
	DepartmentName string                `json:"department_name" gorm:"size:100;not null" binding:"required"`
	CategoryName   string                `json:"category_name,omitempty" gorm:"size:100"`
	Note           string                `json:"note,omitempty" gorm:"type:text"`
	Status         CategoryRequestStatus `json:"status" gorm:"size:20;not null;default:'pending'"`
	AdminResponse  string                `json:"admin_response,omitempty" gorm:"type:text"`
	CreatedAt      time.Time             `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time             `json:"updated_at" gorm:"autoUpdateTime"`

	// Populated for admin-portal's list view only — not a persisted column.
	SellerName string `json:"seller_name,omitempty" gorm:"-"`
	ShopName   string `json:"shop_name,omitempty" gorm:"-"`
}
