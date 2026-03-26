package domain

type Job struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	Title       string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text;not null"`
	Company     string `gorm:"type:varchar(255);not null"`
	Location    string `gorm:"type:varchar(255);not null"`
	Salary      uint   `gorm:"not null"`
	CategoryID  uint   `gorm:"not null"`
	UserID      uint   `gorm:"not null"`
	Category    JobCategory
	PostedDate  string `gorm:"type:varchar(50);not null"`
	ExpiryDate  string `gorm:"type:varchar(50);not null"`
	LocationID  uint   `gorm:"not null"`
}
type JobCategory struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	Name             string `gorm:"type:varchar(100);not null;unique"`
	JobSubCategories []JobSubCategory
	CategoryID       uint
	ParentID         *uint
	Jobs             []Job `gorm:"foreignKey:CategoryID"`
}
type JobSubCategory struct {
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	Name       string `gorm:"type:varchar(100);not null;unique"`
	CategoryID uint   `gorm:"not null"`
}
type JobLocation struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(100);not null;unique"`
}
type JobFilter struct {
	ID     uint   `gorm:"primaryKey;autoIncrement"`
	Name   string `gorm:"type:varchar(100);not null;unique"`
	Type   string `gorm:"type:varchar(50);not null"`
	Values string `gorm:"type:text;not null"` // Comma-separated values
}
type JobCategoryFilter struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	CategoryID uint `gorm:"not null"`
	FilterID   uint `gorm:"not null"`
}
type JobCategoryLocation struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	CategoryID uint `gorm:"not null"`
	LocationID uint `gorm:"not null"`
}
type JobSearchResult struct {
	JobID      uint
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
