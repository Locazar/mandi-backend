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
