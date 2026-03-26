package repository

import (
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type brandDatabase struct {
	DB *gorm.DB
}

func NewBrandDatabaseRepository(db *gorm.DB) interfaces.BrandRepository {
	return &brandDatabase{
		DB: db,
	}
}

func (c *brandDatabase) IsExist(brand domain.Brand) (bool, error) {

	res := c.DB.Where("name = ?", brand.Name).Find(&brand)
	if res.Error != nil {
		return false, res.Error
	}

	return res.RowsAffected != 0, nil
}

func (c *brandDatabase) Save(brand domain.Brand) (domain.Brand, error) {

	err := c.DB.Create(&brand).Error

	return brand, err
}

func (c *brandDatabase) Update(brand domain.Brand) error {

	return c.DB.Where("id = ?", brand.ID).Updates(&brand).Error
}

func (c *brandDatabase) FindAll(pagination request.Pagination) (brands []domain.Brand, err error) {

	err = c.DB.Limit(int(pagination.Limit)).Offset(int(pagination.Offset)).Find(&brands).Error

	return
}

func (c *brandDatabase) FindOne(brandID uint) (brand domain.Brand, err error) {

	err = c.DB.Where("id = ?", brandID).First(&brand).Error

	return
}

func (c *brandDatabase) Delete(brandID uint) error {

	return c.DB.Where("id = ?", brandID).Delete(&domain.Brand{}).Error
}
