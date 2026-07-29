package usecase

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

// LanguageUseCase serves the selectable languages list.
type LanguageUseCase struct {
	repo repo.LanguageRepository
}

func NewLanguageUseCase(repo repo.LanguageRepository) *LanguageUseCase {
	return &LanguageUseCase{repo: repo}
}

func (u *LanguageUseCase) ListActive(ctx context.Context) ([]response.LanguageItem, error) {
	langs, err := u.repo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]response.LanguageItem, 0, len(langs))
	for _, l := range langs {
		items = append(items, response.LanguageItem{
			Code:       l.Code,
			Name:       l.Name,
			NativeName: l.NativeName,
		})
	}
	return items, nil
}
