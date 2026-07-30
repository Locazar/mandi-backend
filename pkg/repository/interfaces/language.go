package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// LanguageRepository reads and manages the selectable languages list.
type LanguageRepository interface {
	// ListActive returns only active languages, ordered — used by the client
	// onboarding picker.
	ListActive(ctx context.Context) ([]domain.Language, error)

	// Admin management.
	ListAll(ctx context.Context) ([]domain.Language, error)
	GetByID(ctx context.Context, id string) (domain.Language, error)
	GetByCode(ctx context.Context, code string) (domain.Language, error)
	Create(ctx context.Context, lang *domain.Language) error
	Update(ctx context.Context, lang *domain.Language) error
	Delete(ctx context.Context, id string) error
}
