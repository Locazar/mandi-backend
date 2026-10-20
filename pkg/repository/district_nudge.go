package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type districtNudgeRepository struct {
	db *gorm.DB
}

func NewDistrictNudgeRepository(db *gorm.DB) interfaces.DistrictNudgeRepository {
	return &districtNudgeRepository{db: db}
}

func defaultDistrictNudgeTemplate() domain.DistrictNudgeTemplate {
	return domain.DistrictNudgeTemplate{
		ID:    domain.DistrictNudgeTemplateID,
		Title: "🚀 {{district_name}} shoppers are noticing — is your shop next?",
		Body: "Hi {{target_merchant_name}}, big moves in {{district_name}}! {{top_shop_name}} just added " +
			"{{product_count}} new products and is getting noticed by shoppers nearby right now. More products " +
			"means more chances to show up when a customer searches — don't let them win this week's attention. " +
			"Update your shop and add your products today.\n\n👉 See what they're doing: {{top_shop_link}}",
		// "Add Product" tab — see admin-portal's action-link.ts (home_add_product).
		Route: "/home?tab=2",
	}
}

func (r *districtNudgeRepository) ensureTables(ctx context.Context) error {
	migrator := r.db.Migrator()
	for _, model := range []interface{}{
		&domain.DistrictNudgeTemplate{},
		&domain.DistrictNudgeSettings{},
		&domain.DistrictNudgeSent{},
	} {
		if !migrator.HasTable(model) {
			if err := r.db.AutoMigrate(model); err != nil {
				return err
			}
		}
	}

	var templateCount int64
	if err := r.db.WithContext(ctx).Model(&domain.DistrictNudgeTemplate{}).Count(&templateCount).Error; err != nil {
		return err
	}
	if templateCount == 0 {
		if err := r.db.WithContext(ctx).Create(&[]domain.DistrictNudgeTemplate{defaultDistrictNudgeTemplate()}).Error; err != nil {
			return err
		}
	}

	var settingsCount int64
	if err := r.db.WithContext(ctx).Model(&domain.DistrictNudgeSettings{}).Count(&settingsCount).Error; err != nil {
		return err
	}
	if settingsCount == 0 {
		defaults := domain.DistrictNudgeSettings{
			ID:       domain.DistrictNudgeSettingsID,
			Enabled:  false,
			State:    "Rajasthan",
			RadiusKm: 100,
		}
		if err := r.db.WithContext(ctx).Create(&defaults).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *districtNudgeRepository) GetTemplate(ctx context.Context) (domain.DistrictNudgeTemplate, error) {
	if err := r.ensureTables(ctx); err != nil {
		return domain.DistrictNudgeTemplate{}, err
	}
	var tmpl domain.DistrictNudgeTemplate
	err := r.db.WithContext(ctx).Where("id = ?", domain.DistrictNudgeTemplateID).First(&tmpl).Error
	return tmpl, err
}

func (r *districtNudgeRepository) SaveTemplate(ctx context.Context, tmpl domain.DistrictNudgeTemplate) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}
	tmpl.ID = domain.DistrictNudgeTemplateID
	return r.db.WithContext(ctx).Model(&domain.DistrictNudgeTemplate{}).
		Where("id = ?", domain.DistrictNudgeTemplateID).
		Updates(map[string]interface{}{
			"title":     tmpl.Title,
			"body":      tmpl.Body,
			"image_url": tmpl.ImageURL,
			"route":     tmpl.Route,
		}).Error
}

func (r *districtNudgeRepository) GetSettings(ctx context.Context) (domain.DistrictNudgeSettings, error) {
	if err := r.ensureTables(ctx); err != nil {
		return domain.DistrictNudgeSettings{}, err
	}
	var settings domain.DistrictNudgeSettings
	err := r.db.WithContext(ctx).Where("id = ?", domain.DistrictNudgeSettingsID).First(&settings).Error
	return settings, err
}

func (r *districtNudgeRepository) SaveSettings(ctx context.Context, settings domain.DistrictNudgeSettings) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}
	settings.ID = domain.DistrictNudgeSettingsID
	return r.db.WithContext(ctx).Model(&domain.DistrictNudgeSettings{}).
		Where("id = ?", domain.DistrictNudgeSettingsID).
		Updates(map[string]interface{}{
			"enabled":   settings.Enabled,
			"state":     settings.State,
			"radius_km": settings.RadiusKm,
		}).Error
}

// FindCandidates self-joins active shops against active shops: for each
// target shop t, the strongest (max product count) other active shop b
// within radiusKm that has MORE products than t, and hasn't already been
// sent for this exact (t, b) pair. DISTINCT ON picks exactly one best peer
// per target. The Haversine distance formula matches search.go's existing
// geo-radius query — same constant, same shape — rather than introducing a
// second way of computing "nearby" in this codebase.
func (r *districtNudgeRepository) FindCandidates(ctx context.Context, state string, radiusKm float64) ([]domain.DistrictNudgeCandidate, error) {
	if err := r.ensureTables(ctx); err != nil {
		return nil, err
	}
	stateFilter := "%"
	if state != "" {
		stateFilter = state
	}
	var candidates []domain.DistrictNudgeCandidate
	err := r.db.WithContext(ctx).Raw(`
		WITH candidates AS (
			SELECT sd.id, sd.shop_name, sd.owner_name, sd.city, sd.state,
				sd.latitude, sd.longitude,
				(SELECT COUNT(*) FROM product_items pi WHERE pi.shop_id = sd.id) AS product_count
			FROM shop_details sd
			WHERE sd.shop_status = 'active'
				AND sd.state ILIKE $1
				AND sd.latitude <> 0 AND sd.longitude <> 0
		)
		SELECT DISTINCT ON (t.id)
			t.id AS target_shop_id, t.shop_name AS target_shop_name,
			t.owner_name AS target_owner_name, t.city AS target_city,
			b.id AS top_shop_id, b.shop_name AS top_shop_name, b.city AS top_shop_city,
			b.product_count AS top_product_count
		FROM candidates t
		JOIN candidates b ON b.id <> t.id
			AND b.product_count > t.product_count
			AND (6371 * acos(
				cos(radians(t.latitude)) * cos(radians(b.latitude)) *
				cos(radians(b.longitude) - radians(t.longitude)) +
				sin(radians(t.latitude)) * sin(radians(b.latitude))
			)) <= $2
			AND NOT EXISTS (
				SELECT 1 FROM district_nudge_sents s
				WHERE s.target_shop_id = t.id AND s.top_shop_id = b.id
			)
		ORDER BY t.id, b.product_count DESC, b.id
	`, stateFilter, radiusKm).Scan(&candidates).Error
	return candidates, err
}

func (r *districtNudgeRepository) MarkSent(ctx context.Context, targetShopID, topShopID string) error {
	sent := domain.DistrictNudgeSent{TargetShopID: targetShopID, TopShopID: topShopID}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&sent).Error
}
