package repository

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type subscriptionExpiredNudgeRepository struct{ db *gorm.DB }

func NewSubscriptionExpiredNudgeRepository(db *gorm.DB) interfaces.SubscriptionExpiredNudgeRepository {
	return &subscriptionExpiredNudgeRepository{db: db}
}

func defaultExpiredNudgeTemplate() domain.SubscriptionExpiredNudgeTemplate {
	return domain.SubscriptionExpiredNudgeTemplate{
		ID:    domain.SubscriptionExpiredNudgeTemplateID,
		Title: "Aapki shop abhi band hai! 🔴",
		Body:  "Aapka Locazar subscription {{days_expired}} din pehle expire ho gaya. Abhi renew karein aur apni shop wapas live karein!",
	}
}

func (r *subscriptionExpiredNudgeRepository) ensureTables(ctx context.Context) error {
	for _, model := range []interface{}{
		&domain.SubscriptionExpiredNudgeSettings{},
		&domain.SubscriptionExpiredNudgeTemplate{},
		&domain.SubscriptionExpiredNudgeSent{},
	} {
		if err := r.db.AutoMigrate(model); err != nil {
			return err
		}
	}

	var settingsCount int64
	if err := r.db.WithContext(ctx).Model(&domain.SubscriptionExpiredNudgeSettings{}).Count(&settingsCount).Error; err != nil {
		return err
	}
	if settingsCount == 0 {
		defaults := domain.SubscriptionExpiredNudgeSettings{
			ID: domain.SubscriptionExpiredNudgeSettingsID, Enabled: false,
			FireHourIST: 9, FireMinuteIST: 0,
		}
		if err := r.db.WithContext(ctx).Create(&defaults).Error; err != nil {
			return err
		}
	}

	var tmplCount int64
	if err := r.db.WithContext(ctx).Model(&domain.SubscriptionExpiredNudgeTemplate{}).Count(&tmplCount).Error; err != nil {
		return err
	}
	if tmplCount == 0 {
		tmpl := defaultExpiredNudgeTemplate()
		if err := r.db.WithContext(ctx).Create(&tmpl).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *subscriptionExpiredNudgeRepository) GetSettings(ctx context.Context) (domain.SubscriptionExpiredNudgeSettings, error) {
	if err := r.ensureTables(ctx); err != nil {
		return domain.SubscriptionExpiredNudgeSettings{}, err
	}
	var s domain.SubscriptionExpiredNudgeSettings
	err := r.db.WithContext(ctx).Where("id = ?", domain.SubscriptionExpiredNudgeSettingsID).First(&s).Error
	return s, err
}

func (r *subscriptionExpiredNudgeRepository) SaveSettings(ctx context.Context, settings domain.SubscriptionExpiredNudgeSettings) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&domain.SubscriptionExpiredNudgeSettings{}).
		Where("id = ?", domain.SubscriptionExpiredNudgeSettingsID).
		Updates(map[string]interface{}{
			"enabled":         settings.Enabled,
			"fire_hour_ist":   settings.FireHourIST,
			"fire_minute_ist": settings.FireMinuteIST,
		}).Error
}

func (r *subscriptionExpiredNudgeRepository) GetTemplate(ctx context.Context) (domain.SubscriptionExpiredNudgeTemplate, error) {
	if err := r.ensureTables(ctx); err != nil {
		return domain.SubscriptionExpiredNudgeTemplate{}, err
	}
	var t domain.SubscriptionExpiredNudgeTemplate
	err := r.db.WithContext(ctx).Where("id = ?", domain.SubscriptionExpiredNudgeTemplateID).First(&t).Error
	return t, err
}

func (r *subscriptionExpiredNudgeRepository) SaveTemplate(ctx context.Context, tmpl domain.SubscriptionExpiredNudgeTemplate) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&domain.SubscriptionExpiredNudgeTemplate{}).
		Where("id = ?", domain.SubscriptionExpiredNudgeTemplateID).
		Updates(map[string]interface{}{
			"title":     tmpl.Title,
			"body":      tmpl.Body,
			"image_url": tmpl.ImageURL,
			"route":     tmpl.Route,
		}).Error
}

// FindAlreadyExpired returns sellers whose subscription has expired and have
// not been nudged today. Sellers who have since renewed are excluded.
// DISTINCT ON (us.user_id) prevents duplicates; ORDER BY end_date DESC picks
// the most recently expired subscription when there are several.
func (r *subscriptionExpiredNudgeRepository) FindAlreadyExpired(ctx context.Context) ([]domain.SubscriptionExpiredNudgeCandidate, error) {
	if err := r.ensureTables(ctx); err != nil {
		return nil, err
	}
	var candidates []domain.SubscriptionExpiredNudgeCandidate
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (us.user_id)
		       us.user_id,
		       COALESCE(sd.shop_name, '')  AS shop_name,
		       COALESCE(sd.owner_name, '') AS owner_name,
		       us.end_date
		FROM user_subscriptions us
		LEFT JOIN shop_details sd ON sd.admin_id = us.user_id
		WHERE us.end_date < NOW()
		  AND NOT EXISTS (
		      SELECT 1 FROM user_subscriptions active
		      WHERE active.user_id = us.user_id
		        AND active.is_active = true
		        AND active.end_date > NOW()
		  )
		  AND NOT EXISTS (
		      SELECT 1 FROM subscription_expired_nudge_sents s
		      WHERE s.user_id = us.user_id
		        AND s.sent_date = TO_CHAR(CURRENT_DATE, 'YYYY-MM-DD')
		  )
		ORDER BY us.user_id, us.end_date DESC
	`).Scan(&candidates).Error
	return candidates, err
}

func (r *subscriptionExpiredNudgeRepository) MarkSent(ctx context.Context, userID string) error {
	sent := domain.SubscriptionExpiredNudgeSent{
		UserID:   userID,
		SentDate: time.Now().Format("2006-01-02"),
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&sent).Error
}
