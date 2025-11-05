package usecase

import (
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type brandUseCase struct {
	brandRepo repoInterface.BrandRepository
}

func NewBrandUseCase(brandRepo repoInterface.BrandRepository) interfaces.BrandUseCase {
	return &brandUseCase{
		brandRepo: brandRepo,
	}
}

func (b *brandUseCase) Save(brand domain.Brand) (domain.Brand, error) {

	alreadyExist, err := b.brandRepo.IsExist(brand)
	if err != nil {
		return domain.Brand{}, utils.PrependMessageToError(err, "failed to check brand name already exist")
	}

	if alreadyExist {
		return domain.Brand{}, ErrBrandAlreadyExist
	}

	brand, err = b.brandRepo.Save(brand)
	if err != nil {
		return domain.Brand{}, utils.PrependMessageToError(err, "failed to save brand on db")
	}

	return brand, nil
}

func (b *brandUseCase) Update(brand domain.Brand) error {

	err := b.brandRepo.Update(brand)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to update brand on db")
	}

	return nil
}

func (b *brandUseCase) FindAll(pagination request.Pagination) ([]domain.Brand, error) {

	brands, err := b.brandRepo.FindAll(pagination)

	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to find all brands from db")
	}

	return brands, nil
}

func (b *brandUseCase) FindOne(brandID uint) (domain.Brand, error) {

	brand, err := b.brandRepo.FindOne(brandID)
	if err != nil {
		return domain.Brand{}, utils.PrependMessageToError(err, "failed to find brand from db")
	}

	return brand, nil
}

func (b *brandUseCase) Delete(brandID uint) error {

	err := b.brandRepo.Delete(brandID)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to delete brands from db")
	}

	return nil
}
