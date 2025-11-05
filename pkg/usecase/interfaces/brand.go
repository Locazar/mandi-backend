package interfaces

import (
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type BrandUseCase interface {
	Save(brand domain.Brand) (domain.Brand, error)
	Update(brand domain.Brand) error
	FindAll(pagination request.Pagination) ([]domain.Brand, error)
	FindOne(brandID uint) (domain.Brand, error)
	Delete(brandID uint) error
}
