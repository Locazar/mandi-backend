package domain

import "time"

// SellerJobPosting is a job requirement posted by a seller from the seller-app.
// It is completely separate from the generic job-marketplace Job model.
type SellerJobPosting struct {
	ID            string     `gorm:"primaryKey;type:varchar(32)"          json:"id"`
	ShopID        *string    `gorm:"type:varchar(32)"                     json:"shop_id,omitempty"`
	ShopName      string     `gorm:"type:varchar(255);not null"           json:"shop_name"`
	Department    string     `gorm:"type:varchar(255);not null"           json:"department"`
	PersonsNeeded int        `gorm:"not null;default:1"                   json:"persons_needed"`
	AgeMin        *int       `                                            json:"age_min,omitempty"`
	AgeMax        *int       `                                            json:"age_max,omitempty"`
	City          string     `gorm:"type:varchar(255);not null;default:''" json:"city"`
	State         string     `gorm:"type:varchar(255);not null;default:''" json:"state"`
	ShowPhone     bool       `gorm:"not null;default:true"                json:"show_phone"`
	Phone         *string    `gorm:"type:varchar(50)"                     json:"phone,omitempty"`
	IsActive      bool       `gorm:"not null;default:true"                json:"is_active"`
	CreatedAt     time.Time  `                                            json:"created_at"`
	UpdatedAt     time.Time  `                                            json:"updated_at"`
}

type Job struct {
	ID          string `gorm:"primaryKey;type:varchar(32)"`
	Title       string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text;not null"`
	Company     string `gorm:"type:varchar(255);not null"`
	Location    string `gorm:"type:varchar(255);not null"`
	Salary      uint   `gorm:"not null"`
	CategoryID  string `gorm:"type:varchar(32);not null"`
	UserID      string `gorm:"type:varchar(32);not null"`
	Category    JobCategory
	PostedDate  string `gorm:"type:varchar(50);not null"`
	ExpiryDate  string `gorm:"type:varchar(50);not null"`
	LocationID  string `gorm:"type:varchar(32);not null"`
}
type JobCategory struct {
	ID               string `gorm:"primaryKey;type:varchar(32)"`
	Name             string `gorm:"type:varchar(100);not null;unique"`
	JobSubCategories []JobSubCategory
	CategoryID       string `gorm:"type:varchar(32)"`
	ParentID         *string `gorm:"type:varchar(32)"`
	Jobs             []Job `gorm:"foreignKey:CategoryID"`
}
type JobSubCategory struct {
	ID         string `gorm:"primaryKey;type:varchar(32)"`
	Name       string `gorm:"type:varchar(100);not null;unique"`
	CategoryID string `gorm:"type:varchar(32);not null"`
}
type JobLocation struct {
	ID   string `gorm:"primaryKey;type:varchar(32)"`
	Name string `gorm:"type:varchar(100);not null;unique"`
}
type JobFilter struct {
	ID     string `gorm:"primaryKey;type:varchar(32)"`
	Name   string `gorm:"type:varchar(100);not null;unique"`
	Type   string `gorm:"type:varchar(50);not null"`
	Values string `gorm:"type:text;not null"` // Comma-separated values
}
type JobCategoryFilter struct {
	ID         string `gorm:"primaryKey;type:varchar(32)"`
	CategoryID string `gorm:"type:varchar(32);not null"`
	FilterID   string `gorm:"type:varchar(32);not null"`
}
type JobCategoryLocation struct {
	ID         string `gorm:"primaryKey;type:varchar(32)"`
	CategoryID string `gorm:"type:varchar(32);not null"`
	LocationID string `gorm:"type:varchar(32);not null"`
}
type JobSearchResult struct {
	JobID      string
	Title      string
	Company    string
	Location   string
	Salary     uint
	PostedDate string
}
type JobCategoryWithSubCategories struct {
	Category      JobCategory
	SubCategories []JobSubCategory
}
