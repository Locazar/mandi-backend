package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// VerificationNotificationRepository stores the admin-editable copy for
// shop document-verification pushes (see domain.VerificationNotificationTemplate).
type VerificationNotificationRepository interface {
	GetTemplates(ctx context.Context) ([]domain.VerificationNotificationTemplate, error)
	GetTemplate(ctx context.Context, key string) (domain.VerificationNotificationTemplate, error)
	SaveTemplate(ctx context.Context, tmpl domain.VerificationNotificationTemplate) error
}
