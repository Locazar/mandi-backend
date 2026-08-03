package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type languageDatabase struct {
	DB *gorm.DB
}

func NewLanguageRepository(db *gorm.DB) repo.LanguageRepository {
	return &languageDatabase{DB: db}
}

func (c *languageDatabase) ListActive(ctx context.Context) ([]domain.Language, error) {
	languages := []domain.Language{}
	err := c.DB.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&languages).Error
	return languages, err
}

func (c *languageDatabase) ListAll(ctx context.Context) ([]domain.Language, error) {
	languages := []domain.Language{}
	err := c.DB.WithContext(ctx).
		Order("sort_order ASC, name ASC").
		Find(&languages).Error
	return languages, err
}

func (c *languageDatabase) GetByID(ctx context.Context, id string) (domain.Language, error) {
	var lang domain.Language
	err := c.DB.WithContext(ctx).Where("id = ?", id).First(&lang).Error
	return lang, err
}

func (c *languageDatabase) GetByCode(ctx context.Context, code string) (domain.Language, error) {
	var lang domain.Language
	err := c.DB.WithContext(ctx).Where("code = ?", code).First(&lang).Error
	return lang, err
}

func (c *languageDatabase) Create(ctx context.Context, lang *domain.Language) error {
	return c.DB.WithContext(ctx).Create(lang).Error
}

func (c *languageDatabase) Update(ctx context.Context, lang *domain.Language) error {
	// Full-row save; the caller loads-then-mutates so unset columns aren't zeroed.
	return c.DB.WithContext(ctx).Save(lang).Error
}

func (c *languageDatabase) Delete(ctx context.Context, id string) error {
	return c.DB.WithContext(ctx).Where("id = ?", id).Delete(&domain.Language{}).Error
}
