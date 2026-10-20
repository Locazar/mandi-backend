package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type verificationNotificationRepository struct {
	db *gorm.DB
}

func NewVerificationNotificationRepository(db *gorm.DB) interfaces.VerificationNotificationRepository {
	return &verificationNotificationRepository{db: db}
}

// defaultVerificationTemplates seeds copy equivalent to the previous
// hardcoded Go strings (shopVerificationMessage / ApproveShop / RejectShop),
// so a fresh deploy sends the same messages as before until an admin edits
// them. Where the old code branched on sub-combinations of the four checks
// (e.g. "business verified, identity pending" vs the reverse), that's
// flattened here into {{business_doc_status}}/{{identity_doc_status}}
// placeholders in one body — simpler to keep admin-editable than embedding
// conditional branches in a plain title/body template.
func defaultVerificationTemplates() []domain.VerificationNotificationTemplate {
	return []domain.VerificationNotificationTemplate{
		{
			Key:   domain.VerifKeyFullyVerified,
			Title: "🎉 Welcome to Locazar — your shop is fully verified!",
			Body: "Congratulations! Your shop photo, shop address, business document and identity document are all " +
				"verified. Your shop is now LIVE and visible to customers. Welcome aboard — happy selling!",
			Route: "/kyc",
		},
		{
			Key:   domain.VerifKeyLivePartial,
			Title: "✅ Your shop is verified and live!",
			Body: "Great news! Your shop photo and shop address are verified, so your shop is now LIVE and visible " +
				"to customers on Locazar. Business document: {{business_doc_status}}. Identity document: " +
				"{{identity_doc_status}}. Welcome aboard!",
			Route: "/kyc",
		},
		{
			Key:   domain.VerifKeyPendingReview,
			Title: "⚠️ Action needed — shop verification pending",
			Body: "Your shop is currently under review and not visible to customers yet. Status — Shop photo: " +
				"{{photo_status}}, Shop address: {{address_status}}, Business document: {{business_doc_status}}, " +
				"Identity document: {{identity_doc_status}}. Your shop will go live once both your shop photo and " +
				"shop address are verified — please make sure they are clear and valid.",
			Route: "/kyc",
		},
		{
			Key:   domain.VerifKeyApproved,
			Title: "🎉 Congratulations, you're live!",
			Body:  "Your shop has been approved and is now visible to customers on Locazar.",
			Route: "/kyc",
		},
		{
			Key:   domain.VerifKeyRejected,
			Title: "Shop verification declined",
			Body:  "Your shop verification was declined: {{remark}}Please review the feedback and resubmit.",
			Route: "/kyc",
		},
	}
}

// ensureTables lazily migrates and seeds — same pattern as
// onboardingNudgeRepository.ensureTables, so no standalone migration step is
// needed before this feature works.
func (r *verificationNotificationRepository) ensureTables(ctx context.Context) error {
	if !r.db.Migrator().HasTable(&domain.VerificationNotificationTemplate{}) {
		if err := r.db.AutoMigrate(&domain.VerificationNotificationTemplate{}); err != nil {
			return err
		}
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.VerificationNotificationTemplate{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		if err := r.db.WithContext(ctx).Create(defaultVerificationTemplates()).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *verificationNotificationRepository) GetTemplates(ctx context.Context) ([]domain.VerificationNotificationTemplate, error) {
	if err := r.ensureTables(ctx); err != nil {
		return nil, err
	}
	var templates []domain.VerificationNotificationTemplate
	err := r.db.WithContext(ctx).Order("key").Find(&templates).Error
	return templates, err
}

// GetTemplate loads one template by key. Falls back to its hardcoded default
// (never an error) if the row is somehow missing — a best-effort notification
// send must never fail because a template row wasn't found.
func (r *verificationNotificationRepository) GetTemplate(ctx context.Context, key string) (domain.VerificationNotificationTemplate, error) {
	if err := r.ensureTables(ctx); err != nil {
		return r.defaultFor(key), nil
	}
	var tmpl domain.VerificationNotificationTemplate
	if err := r.db.WithContext(ctx).First(&tmpl, "key = ?", key).Error; err != nil {
		return r.defaultFor(key), nil
	}
	return tmpl, nil
}

func (r *verificationNotificationRepository) defaultFor(key string) domain.VerificationNotificationTemplate {
	for _, d := range defaultVerificationTemplates() {
		if d.Key == key {
			return d
		}
	}
	return domain.VerificationNotificationTemplate{Key: key}
}

func (r *verificationNotificationRepository) SaveTemplate(ctx context.Context, tmpl domain.VerificationNotificationTemplate) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Model(&domain.VerificationNotificationTemplate{}).
		Where("key = ?", tmpl.Key).
		Updates(map[string]interface{}{
			"title":     tmpl.Title,
			"body":      tmpl.Body,
			"image_url": tmpl.ImageURL,
			"route":     tmpl.Route,
		}).Error
}
