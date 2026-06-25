package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

type ProductRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo interfaces.ProductRepository
}

func (suite *ProductRepositoryTestSuite) SetupTest() {
	// Create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(suite.T(), err)

	// Create the necessary tables for testing
	err = db.Exec(`
		CREATE TABLE departments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			sort_order INTEGER DEFAULT 0,
			is_active BOOLEAN DEFAULT true,
			image_url TEXT,
			icon TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	assert.NoError(suite.T(), err)

	err = db.Exec(`
		CREATE TABLE categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			department_id INTEGER NOT NULL,
			sort_order INTEGER DEFAULT 0,
			is_active BOOLEAN DEFAULT true,
			image_url TEXT,
			icon TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (department_id) REFERENCES departments(id)
		)
	`).Error
	assert.NoError(suite.T(), err)

	err = db.Exec(`
		CREATE TABLE sub_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			department_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			sort_order INTEGER DEFAULT 0,
			is_active BOOLEAN DEFAULT true,
			image_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (department_id) REFERENCES departments(id),
			FOREIGN KEY (category_id) REFERENCES categories(id)
		)
	`).Error
	assert.NoError(suite.T(), err)

	suite.db = db
	suite.repo = NewProductRepository(db, nil)
}

func (suite *ProductRepositoryTestSuite) TearDownTest() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

// Tests for SaveDepartment
func (suite *ProductRepositoryTestSuite) TestSaveDepartment_Success() {
	department := request.Department{
		Name:      "Electronics",
		Slug:      "electronics",
		SortOrder: 1,
		IsActive:  true,
		ImageURL:  "/uploads/departments/electronics.png",
		Icon:      "icon-electronics",
	}

	err := suite.repo.SaveDepartment(context.Background(), department)
	assert.NoError(suite.T(), err)

	// Verify the department was saved
	var savedDept struct {
		ID        int
		Name      string
		Slug      string
		SortOrder int
		IsActive  bool
		ImageURL  string
		Icon      string
	}

	result := suite.db.Raw("SELECT id, name, slug, sort_order, is_active, image_url, icon FROM departments WHERE slug = ?", "electronics").Scan(&savedDept)
	assert.NoError(suite.T(), result.Error)
	assert.Equal(suite.T(), "Electronics", savedDept.Name)
	assert.Equal(suite.T(), "electronics", savedDept.Slug)
	assert.Equal(suite.T(), 1, savedDept.SortOrder)
	assert.Equal(suite.T(), true, savedDept.IsActive)
	assert.Equal(suite.T(), "/uploads/departments/electronics.png", savedDept.ImageURL)
	assert.Equal(suite.T(), "icon-electronics", savedDept.Icon)
}

func (suite *ProductRepositoryTestSuite) TestSaveDepartment_WithMinimalFields() {
	department := request.Department{
		Name:     "Clothing",
		Slug:     "clothing",
		IsActive: false,
		ImageURL: "/uploads/departments/clothing.png",
	}

	err := suite.repo.SaveDepartment(context.Background(), department)
	assert.NoError(suite.T(), err)

	var savedDept struct {
		Name     string
		Slug     string
		IsActive bool
	}

	result := suite.db.Raw("SELECT name, slug, is_active FROM departments WHERE slug = ?", "clothing").Scan(&savedDept)
	assert.NoError(suite.T(), result.Error)
	assert.Equal(suite.T(), "Clothing", savedDept.Name)
	assert.False(suite.T(), savedDept.IsActive)
}

func (suite *ProductRepositoryTestSuite) TestSaveDepartment_MultipleDepartments() {
	departments := []request.Department{
		{
			Name:      "Electronics",
			Slug:      "electronics",
			SortOrder: 1,
			IsActive:  true,
			ImageURL:  "/uploads/departments/electronics.png",
		},
		{
			Name:      "Clothing",
			Slug:      "clothing",
			SortOrder: 2,
			IsActive:  true,
			ImageURL:  "/uploads/departments/clothing.png",
		},
		{
			Name:      "Books",
			Slug:      "books",
			SortOrder: 3,
			IsActive:  true,
			ImageURL:  "/uploads/departments/books.png",
		},
	}

	for _, dept := range departments {
		err := suite.repo.SaveDepartment(context.Background(), dept)
		assert.NoError(suite.T(), err)
	}

	// Verify all departments were saved
	var count int64
	suite.db.Raw("SELECT COUNT(*) FROM departments").Scan(&count)
	assert.Equal(suite.T(), int64(3), count)
}

// Tests for SaveCategory
func (suite *ProductRepositoryTestSuite) TestSaveCategory_Success() {
	// First, create a department
	dept := request.Department{
		Name:     "Electronics",
		Slug:     "electronics",
		IsActive: true,
		ImageURL: "/uploads/departments/electronics.png",
	}
	suite.repo.SaveDepartment(context.Background(), dept)

	// Get the department ID
	var deptID int
	suite.db.Raw("SELECT id FROM departments WHERE slug = ?", "electronics").Scan(&deptID)

	category := request.Category{
		Name:      "Smartphones",
		SortOrder: 1,
		IsActive:  true,
		ImageURL:  "/uploads/categories/smartphones.png",
		Icon:      "icon-smartphone",
	}

	err := suite.repo.SaveCategory(context.Background(), category, string(rune(deptID)))
	assert.NoError(suite.T(), err)

	// Verify the category was saved
	var savedCat struct {
		ID        int
		Name      string
		SortOrder int
		IsActive  bool
		ImageURL  string
		Icon      string
	}

	result := suite.db.Raw("SELECT id, name, sort_order, is_active, image_url, icon FROM categories WHERE name = ?", "Smartphones").Scan(&savedCat)
	assert.NoError(suite.T(), result.Error)
	assert.Equal(suite.T(), "Smartphones", savedCat.Name)
	assert.Equal(suite.T(), 1, savedCat.SortOrder)
	assert.True(suite.T(), savedCat.IsActive)
	assert.Equal(suite.T(), "/uploads/categories/smartphones.png", savedCat.ImageURL)
	assert.Equal(suite.T(), "icon-smartphone", savedCat.Icon)
}

func (suite *ProductRepositoryTestSuite) TestSaveCategory_WithMinimalFields() {
	// Create a department first
	dept := request.Department{
		Name:     "Electronics",
		Slug:     "electronics",
		IsActive: true,
		ImageURL: "/uploads/departments/electronics.png",
	}
	suite.repo.SaveDepartment(context.Background(), dept)

	var deptID int
	suite.db.Raw("SELECT id FROM departments WHERE slug = ?", "electronics").Scan(&deptID)

	category := request.Category{
		Name:     "Laptops",
		IsActive: false,
		ImageURL: "/uploads/categories/laptops.png",
	}

	err := suite.repo.SaveCategory(context.Background(), category, string(rune(deptID)))
	assert.NoError(suite.T(), err)

	var savedCat struct {
		Name     string
		IsActive bool
	}

	result := suite.db.Raw("SELECT name, is_active FROM categories WHERE name = ?", "Laptops").Scan(&savedCat)
	assert.NoError(suite.T(), result.Error)
	assert.Equal(suite.T(), "Laptops", savedCat.Name)
	assert.False(suite.T(), savedCat.IsActive)
}

func (suite *ProductRepositoryTestSuite) TestSaveCategory_MultipleCategoriesInDepartment() {
	// Create a department
	dept := request.Department{
		Name:     "Electronics",
		Slug:     "electronics",
		IsActive: true,
		ImageURL: "/uploads/departments/electronics.png",
	}
	suite.repo.SaveDepartment(context.Background(), dept)

	var deptID int
	suite.db.Raw("SELECT id FROM departments WHERE slug = ?", "electronics").Scan(&deptID)
	deptIDStr := string(rune(deptID))

	categories := []request.Category{
		{
			Name:      "Smartphones",
			SortOrder: 1,
			IsActive:  true,
			ImageURL:  "/uploads/categories/smartphones.png",
		},
		{
			Name:      "Laptops",
			SortOrder: 2,
			IsActive:  true,
			ImageURL:  "/uploads/categories/laptops.png",
		},
		{
			Name:      "Tablets",
			SortOrder: 3,
			IsActive:  true,
			ImageURL:  "/uploads/categories/tablets.png",
		},
	}

	for _, cat := range categories {
		err := suite.repo.SaveCategory(context.Background(), cat, deptIDStr)
		assert.NoError(suite.T(), err)
	}

	// Verify all categories were saved
	var count int64
	suite.db.Raw("SELECT COUNT(*) FROM categories WHERE department_id = ?", deptID).Scan(&count)
	assert.Equal(suite.T(), int64(3), count)
}

// Tests for SaveSubCategory
func (suite *ProductRepositoryTestSuite) TestSaveSubCategory_Success() {
	// Create department and category
	dept := request.Department{
		Name:     "Electronics",
		Slug:     "electronics",
		IsActive: true,
		ImageURL: "/uploads/departments/electronics.png",
	}
	suite.repo.SaveDepartment(context.Background(), dept)

	var deptID int
	suite.db.Raw("SELECT id FROM departments WHERE slug = ?", "electronics").Scan(&deptID)
	deptIDStr := string(rune(deptID))

	category := request.Category{
		Name:     "Smartphones",
		IsActive: true,
		ImageURL: "/uploads/categories/smartphones.png",
	}
	suite.repo.SaveCategory(context.Background(), category, deptIDStr)

	var catID int
	suite.db.Raw("SELECT id FROM categories WHERE name = ?", "Smartphones").Scan(&catID)
	catIDStr := string(rune(catID))

	subCategory := request.SubCategory{
		Name:      "Android Phones",
		SortOrder: 1,
		IsActive:  true,
		ImageURL:  "/uploads/subcategories/android.png",
	}

	err := suite.repo.SaveSubCategory(context.Background(), subCategory, deptIDStr, catIDStr)
	assert.NoError(suite.T(), err)

	// Verify the sub-category was saved
	var savedSubCat struct {
		ID        int
		Name      string
		SortOrder int
		IsActive  bool
		ImageURL  string
	}

	result := suite.db.Raw("SELECT id, name, sort_order, is_active, image_url FROM sub_categories WHERE name = ?", "Android Phones").Scan(&savedSubCat)
	assert.NoError(suite.T(), result.Error)
	assert.Equal(suite.T(), "Android Phones", savedSubCat.Name)
	assert.Equal(suite.T(), 1, savedSubCat.SortOrder)
	assert.True(suite.T(), savedSubCat.IsActive)
	assert.Equal(suite.T(), "/uploads/subcategories/android.png", savedSubCat.ImageURL)
}

func (suite *ProductRepositoryTestSuite) TestSaveSubCategory_WithMinimalFields() {
	// Create department and category
	dept := request.Department{
		Name:     "Electronics",
		Slug:     "electronics",
		IsActive: true,
		ImageURL: "/uploads/departments/electronics.png",
	}
	suite.repo.SaveDepartment(context.Background(), dept)

	var deptID int
	suite.db.Raw("SELECT id FROM departments WHERE slug = ?", "electronics").Scan(&deptID)
	deptIDStr := string(rune(deptID))

	category := request.Category{
		Name:     "Smartphones",
		IsActive: true,
		ImageURL: "/uploads/categories/smartphones.png",
	}
	suite.repo.SaveCategory(context.Background(), category, deptIDStr)

	var catID int
	suite.db.Raw("SELECT id FROM categories WHERE name = ?", "Smartphones").Scan(&catID)
	catIDStr := string(rune(catID))

	subCategory := request.SubCategory{
		Name:     "iPhones",
		IsActive: false,
		ImageURL: "/uploads/subcategories/iphone.png",
	}

	err := suite.repo.SaveSubCategory(context.Background(), subCategory, deptIDStr, catIDStr)
	assert.NoError(suite.T(), err)

	var savedSubCat struct {
		Name     string
		IsActive bool
	}

	result := suite.db.Raw("SELECT name, is_active FROM sub_categories WHERE name = ?", "iPhones").Scan(&savedSubCat)
	assert.NoError(suite.T(), result.Error)
	assert.Equal(suite.T(), "iPhones", savedSubCat.Name)
	assert.False(suite.T(), savedSubCat.IsActive)
}

func (suite *ProductRepositoryTestSuite) TestSaveSubCategory_MultipleSubCategories() {
	// Create department and category
	dept := request.Department{
		Name:     "Electronics",
		Slug:     "electronics",
		IsActive: true,
		ImageURL: "/uploads/departments/electronics.png",
	}
	suite.repo.SaveDepartment(context.Background(), dept)

	var deptID int
	suite.db.Raw("SELECT id FROM departments WHERE slug = ?", "electronics").Scan(&deptID)
	deptIDStr := string(rune(deptID))

	category := request.Category{
		Name:     "Smartphones",
		IsActive: true,
		ImageURL: "/uploads/categories/smartphones.png",
	}
	suite.repo.SaveCategory(context.Background(), category, deptIDStr)

	var catID int
	suite.db.Raw("SELECT id FROM categories WHERE name = ?", "Smartphones").Scan(&catID)
	catIDStr := string(rune(catID))

	subCategories := []request.SubCategory{
		{
			Name:      "Android Phones",
			SortOrder: 1,
			IsActive:  true,
			ImageURL:  "/uploads/subcategories/android.png",
		},
		{
			Name:      "iPhones",
			SortOrder: 2,
			IsActive:  true,
			ImageURL:  "/uploads/subcategories/iphone.png",
		},
		{
			Name:      "Windows Phones",
			SortOrder: 3,
			IsActive:  false,
			ImageURL:  "/uploads/subcategories/windows.png",
		},
	}

	for _, subCat := range subCategories {
		err := suite.repo.SaveSubCategory(context.Background(), subCat, deptIDStr, catIDStr)
		assert.NoError(suite.T(), err)
	}

	// Verify all sub-categories were saved
	var count int64
	suite.db.Raw("SELECT COUNT(*) FROM sub_categories WHERE category_id = ?", catID).Scan(&count)
	assert.Equal(suite.T(), int64(3), count)
}

func (suite *ProductRepositoryTestSuite) TestSaveSubCategory_WithDifferentCategories() {
	// Create department
	dept := request.Department{
		Name:     "Electronics",
		Slug:     "electronics",
		IsActive: true,
		ImageURL: "/uploads/departments/electronics.png",
	}
	suite.repo.SaveDepartment(context.Background(), dept)

	var deptID int
	suite.db.Raw("SELECT id FROM departments WHERE slug = ?", "electronics").Scan(&deptID)
	deptIDStr := string(rune(deptID))

	// Create two categories
	categories := []request.Category{
		{Name: "Smartphones", IsActive: true, ImageURL: "/uploads/categories/smartphones.png"},
		{Name: "Laptops", IsActive: true, ImageURL: "/uploads/categories/laptops.png"},
	}

	var catIDs []int
	for _, cat := range categories {
		suite.repo.SaveCategory(context.Background(), cat, deptIDStr)
	}

	suite.db.Raw("SELECT id FROM categories WHERE department_id = ? ORDER BY id", deptID).Scan(&catIDs)

	// Add sub-categories to different categories
	subCategories := []struct {
		catID  int
		subCat request.SubCategory
	}{
		{
			catID: catIDs[0],
			subCat: request.SubCategory{
				Name:     "Android Phones",
				IsActive: true,
				ImageURL: "/uploads/subcategories/android.png",
			},
		},
		{
			catID: catIDs[1],
			subCat: request.SubCategory{
				Name:     "Gaming Laptops",
				IsActive: true,
				ImageURL: "/uploads/subcategories/gaming.png",
			},
		},
	}

	for _, item := range subCategories {
		err := suite.repo.SaveSubCategory(context.Background(), item.subCat, deptIDStr, string(rune(item.catID)))
		assert.NoError(suite.T(), err)
	}

	// Verify sub-categories in each category
	var smartphoneSubCatCount int64
	suite.db.Raw("SELECT COUNT(*) FROM sub_categories WHERE category_id = ?", catIDs[0]).Scan(&smartphoneSubCatCount)
	assert.Equal(suite.T(), int64(1), smartphoneSubCatCount)

	var laptopSubCatCount int64
	suite.db.Raw("SELECT COUNT(*) FROM sub_categories WHERE category_id = ?", catIDs[1]).Scan(&laptopSubCatCount)
	assert.Equal(suite.T(), int64(1), laptopSubCatCount)
}

func TestProductRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ProductRepositoryTestSuite))
}
