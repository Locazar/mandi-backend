package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// LanguageRepository reads the selectable languages list.
type LanguageRepository interface {
	ListActive(ctx context.Context) ([]domain.Language, error)
}
