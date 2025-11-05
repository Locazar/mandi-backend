package interfaces

import (
<<<<<<< HEAD
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
=======
	"github.com/nikhilnarayanan623/ecommerce-gin-clean-arch/pkg/api/handler/request"
	"github.com/nikhilnarayanan623/ecommerce-gin-clean-arch/pkg/domain"
>>>>>>> b9ab446 (Initial commit)
)

type BrandRepository interface {
	IsExist(brand domain.Brand) (bool, error)
	Save(brand domain.Brand) (domain.Brand, error)
	Update(brand domain.Brand) error
	FindAll(pagination request.Pagination) ([]domain.Brand, error)
	FindOne(brandID uint) (domain.Brand, error)
	Delete(brandID uint) error
}
