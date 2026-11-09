package repository

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type subscriptionExpiryNudgeRepository struct{ db *gorm.DB }

func NewSubscriptionExpiryNudgeRepository(db *gorm.DB) interfaces.SubscriptionExpiryNudgeRepository {
	return &subscriptionExpiryNudgeRepository{db: db}
}

func defaultExpiryNudgeTemplate() domain.SubscriptionExpiryNudgeTemplate {
	return domain.SubscriptionExpiryNudgeTemplate{
		ID:    domain.SubscriptionExpiryNudgeTemplateID,
		Title: "Subscription khatam hone wala hai! ⏰",
		Body:  "Aapka Locazar subscription sirf {{days_left}} din mein expire ho jayega. Abhi renew karein aur apni shop live rakhe customers ke liye!",
	}
}

func (r *subscriptionExpiryNudgeRepository) ensureTables(ctx context.Context) error {
	migrator := r.db.Migrator()
	for _, model := range []interface{}{
		&domain.SubscriptionExpiryNudgeSettings{},
		&domain.SubscriptionExpiryNudgeTemplate{},
		&domain.SubscriptionExpiryNudgeSent{},
	} {
		if !migrator.HasTable(model) {
			if err := r.db.AutoMigrate(model); err != nil {
				return err
			}
		}
	}

	var settingsCount int64
	if err := r.db.WithContext(ctx).Model(&domain.SubscriptionExpiryNudgeSettings{}).Count(&settingsCount).Error; err != nil {
		return err
	}
	if settingsCount == 0 {
		defaults := domain.SubscriptionExpiryNudgeSettings{
			ID: domain.SubscriptionExpiryNudgeSettingsID, Enabled: false, DaysBeforeExpiry: 5,
		}
		if err := r.db.WithContext(ctx).Create(&defaults).Error; err != nil {
			return err
		}
	}

	var tmplCount int64
	if err := r.db.WithContext(ctx).Model(&domain.SubscriptionExpiryNudgeTemplate{}).Count(&tmplCount).Error; err != nil {
		return err
	}
	if tmplCount == 0 {
		tmpl := defaultExpiryNudgeTemplate()
		if err := r.db.WithContext(ctx).Create(&tmpl).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *subscriptionExpiryNudgeRepository) GetSettings(ctx context.Context) (domain.SubscriptionExpiryNudgeSettings, error) {
	if err := r.ensureTables(ctx); err != nil {
		return domain.SubscriptionExpiryNudgeSettings{}, err
	}
	var s domain.SubscriptionExpiryNudgeSettings
	err := r.db.WithContext(ctx).Where("id = ?", domain.SubscriptionExpiryNudgeSettingsID).First(&s).Error
	return s, err
}

func (r *subscriptionExpiryNudgeRepository) SaveSettings(ctx context.Context, settings domain.SubscriptionExpiryNudgeSettings) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&domain.SubscriptionExpiryNudgeSettings{}).
		Where("id = ?", domain.SubscriptionExpiryNudgeSettingsID).
		Updates(map[string]interface{}{
			"enabled":            settings.Enabled,
			"days_before_expiry": settings.DaysBeforeExpiry,
		}).Error
}

func (r *subscriptionExpiryNudgeRepository) GetTemplate(ctx context.Context) (domain.SubscriptionExpiryNudgeTemplate, error) {
	if err := r.ensureTables(ctx); err != nil {
		return domain.SubscriptionExpiryNudgeTemplate{}, err
	}
	var t domain.SubscriptionExpiryNudgeTemplate
	err := r.db.WithContext(ctx).Where("id = ?", domain.SubscriptionExpiryNudgeTemplateID).First(&t).Error
	return t, err
}

func (r *subscriptionExpiryNudgeRepository) SaveTemplate(ctx context.Context, tmpl domain.SubscriptionExpiryNudgeTemplate) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&domain.SubscriptionExpiryNudgeTemplate{}).
		Where("id = ?", domain.SubscriptionExpiryNudgeTemplateID).
		Updates(map[string]interface{}{
			"title":     tmpl.Title,
			"body":      tmpl.Body,
			"image_url": tmpl.ImageURL,
			"route":     tmpl.Route,
		}).Error
}

// FindExpiringSoon returns active subscriptions expiring within daysBeforeExpiry
// calendar days that have not already been nudged today. The $1 multiplier on
// INTERVAL '1 day' is a standard PostgreSQL pattern for parameterised intervals.
func (r *subscriptionExpiryNudgeRepository) FindExpiringSoon(ctx context.Context, daysBeforeExpiry int) ([]domain.SubscriptionExpiryNudgeCandidate, error) {
	if err := r.ensureTables(ctx); err != nil {
		return nil, err
	}
	var candidates []domain.SubscriptionExpiryNudgeCandidate
	// DISTINCT ON (us.user_id) prevents a seller with multiple active
	// subscriptions, or multiple shop_details rows, from appearing more than
	// once and receiving duplicate notifications. ORDER BY end_date ASC picks
	// the soonest-expiring subscription when there are several.
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (us.user_id)
		       us.user_id,
		       COALESCE(sd.shop_name, '')  AS shop_name,
		       COALESCE(sd.owner_name, '') AS owner_name,
		       us.end_date
		FROM user_subscriptions us
		LEFT JOIN shop_details sd ON sd.admin_id = us.user_id
		WHERE us.is_active = true
		  AND us.end_date > NOW()
		  AND us.end_date <= NOW() + ($1 * INTERVAL '1 day')
		  AND NOT EXISTS (
		      SELECT 1 FROM subscription_expiry_nudge_sents s
		      WHERE s.user_id = us.user_id
		        AND s.sent_date = TO_CHAR(CURRENT_DATE, 'YYYY-MM-DD')
		  )
		ORDER BY us.user_id, us.end_date ASC
	`, daysBeforeExpiry).Scan(&candidates).Error
	return candidates, err
}

func (r *subscriptionExpiryNudgeRepository) MarkSent(ctx context.Context, userID string) error {
	sent := domain.SubscriptionExpiryNudgeSent{
		UserID:   userID,
		SentDate: time.Now().Format("2006-01-02"),
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&sent).Error
}
