package interfaces

import "context"

type AlertTemplateUseCase interface {
	// Admin CRUD
	ListTemplates(ctx context.Context) ([]map[string]interface{}, error)
	CreateTemplate(ctx context.Context, req map[string]interface{}) error
	UpdateTemplate(ctx context.Context, key string, req map[string]interface{}) error
	DeleteTemplate(ctx context.Context, key string) error
	ToggleTemplate(ctx context.Context, key string, enabled bool) error
	SeedDefaults(ctx context.Context) error

	// Seller-facing
	GetTemplateForSeller(ctx context.Context, sellerID string, key string) (map[string]interface{}, error)
	GetFlow(ctx context.Context, sellerID string, flowKey string) (map[string]interface{}, error)
	CompleteStep(ctx context.Context, sellerID string, flowKey string, stepNumber int) error
}
