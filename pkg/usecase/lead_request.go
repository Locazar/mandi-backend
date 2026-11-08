package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repointerfacespkg "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	ucinterfacespkg "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type leadRequestUseCase struct {
	repo repointerfacespkg.LeadRequestRepository
}

func NewLeadRequestUseCase(repo repointerfacespkg.LeadRequestRepository) ucinterfacespkg.LeadRequestUseCase {
	return &leadRequestUseCase{repo: repo}
}

func (uc *leadRequestUseCase) Submit(ctx context.Context, phone, requirement, city, source string) (*domain.LeadRequest, error) {
	phone = strings.TrimSpace(phone)
	requirement = strings.TrimSpace(requirement)
	city = strings.TrimSpace(city)

	if phone == "" {
		return nil, fmt.Errorf("phone is required")
	}
	if requirement == "" {
		return nil, fmt.Errorf("requirement is required")
	}

	lead := &domain.LeadRequest{
		Phone:       phone,
		Requirement: requirement,
		City:        city,
		Source:      source,
		Status:      "new",
	}
	if err := uc.repo.Create(ctx, lead); err != nil {
		return nil, err
	}
	return lead, nil
}

func (uc *leadRequestUseCase) List(ctx context.Context, city, status string, limit, offset int) ([]domain.LeadRequest, int64, error) {
	return uc.repo.List(ctx, city, status, limit, offset)
}

func (uc *leadRequestUseCase) UpdateStatus(ctx context.Context, id uint, status, adminNote string) error {
	return uc.repo.UpdateStatus(ctx, id, status, adminNote)
}
