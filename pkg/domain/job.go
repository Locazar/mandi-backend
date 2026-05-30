package domain

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
