package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
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

// ---- Admin management ----

// ListAll returns every language (active + inactive) with full fields for the
// admin management table.
func (u *LanguageUseCase) ListAll(ctx context.Context) ([]domain.Language, error) {
	return u.repo.ListAll(ctx)
}

// Create validates and inserts a new language. Code must be unique.
func (u *LanguageUseCase) Create(ctx context.Context, req request.LanguageRequest) (domain.Language, error) {
	code := strings.ToLower(strings.TrimSpace(req.Code))
	name := strings.TrimSpace(req.Name)
	nativeName := strings.TrimSpace(req.NativeName)
	if code == "" || name == "" {
		return domain.Language{}, errors.New("code and name are required")
	}
	if nativeName == "" {
		nativeName = name
	}

	// Reject duplicate codes with a clear message instead of a raw DB error.
	if _, err := u.repo.GetByCode(ctx, code); err == nil {
		return domain.Language{}, errors.New("a language with this code already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Language{}, err
	}

	lang := domain.Language{
		ID:         domain.NewID(domain.PrefixLanguage),
		Code:       code,
		Name:       name,
		NativeName: nativeName,
		IsActive:   true,
	}
	if req.SortOrder != nil {
		lang.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		lang.IsActive = *req.IsActive
	}
	if err := u.repo.Create(ctx, &lang); err != nil {
		return domain.Language{}, err
	}
	return lang, nil
}

// Update mutates only the fields present in the request (load-then-patch).
func (u *LanguageUseCase) Update(ctx context.Context, id string, req request.LanguageRequest) (domain.Language, error) {
	lang, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Language{}, errors.New("language not found")
		}
		return domain.Language{}, err
	}

	if code := strings.ToLower(strings.TrimSpace(req.Code)); code != "" && code != lang.Code {
		// Guard uniqueness when the code changes.
		if existing, gErr := u.repo.GetByCode(ctx, code); gErr == nil && existing.ID != id {
			return domain.Language{}, errors.New("a language with this code already exists")
		} else if gErr != nil && !errors.Is(gErr, gorm.ErrRecordNotFound) {
			return domain.Language{}, gErr
		}
		lang.Code = code
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		lang.Name = name
	}
	if nativeName := strings.TrimSpace(req.NativeName); nativeName != "" {
		lang.NativeName = nativeName
	}
	if req.SortOrder != nil {
		lang.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		lang.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, &lang); err != nil {
		return domain.Language{}, err
	}
	return lang, nil
}

// Delete removes a language by id.
func (u *LanguageUseCase) Delete(ctx context.Context, id string) error {
	if _, err := u.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("language not found")
		}
		return err
	}
	return u.repo.Delete(ctx, id)
}
