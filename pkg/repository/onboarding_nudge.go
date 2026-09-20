package repository

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type onboardingNudgeRepository struct {
	db *gorm.DB
}

func NewOnboardingNudgeRepository(db *gorm.DB) interfaces.OnboardingNudgeRepository {
	return &onboardingNudgeRepository{db: db}
}

// defaultTemplates seed sensible, ready-to-send copy for all four slots — an
// admin only needs to open the page and edit if the default doesn't fit.
func defaultTemplates() []domain.OnboardingNudgeTemplate {
	return []domain.OnboardingNudgeTemplate{
		{
			Key:   domain.NudgeAddProducts,
			Title: "Add your first products 🛍️",
			Body:  "Your shop is live, but customers can't buy what isn't listed yet. Add a few products to start getting orders on Locazar.",
			// "Add Product" tab — see admin-portal's action-link.ts (home_add_product).
			Route: "/home?tab=2",
		},
		{
			Key:   domain.NudgeUpdatePhoto,
			Title: "Add a shop photo 📸",
			Body:  "Shops with a clear photo get far more customer visits. Open My Shop → Shop Photo and upload one.",
			// "Store" tab, where shop photo lives (home_store).
			Route: "/home?tab=3",
		},
		{
			Key:   domain.NudgeUpdateAddress,
			Title: "Confirm your shop address 📍",
			Body:  "Nearby customers find you by location — double-check your address and map pin are accurate.",
			Route: "/home?tab=3",
		},
		{
			Key:   domain.NudgeViewShop,
			Title: "See your shop on Locazar 🔗",
			Body:  "Your shop is public now — here's your own link: {{shop_link}}",
			// Share & QR screen (share) — the seller can post the link to social
			// media straight from the tap instead of copying it out of the text.
			Route: "/share",
		},
	}
}

// ensureTables lazily migrates and seeds — same pattern as
// fcmTokenRepository.ensureNotificationDeviceTokenTable, so no standalone
// migration step is needed before this feature works.
func (r *onboardingNudgeRepository) ensureTables(ctx context.Context) error {
	migrator := r.db.Migrator()
	for _, model := range []interface{}{
		&domain.OnboardingNudgeTemplate{},
		&domain.OnboardingNudgeSettings{},
		&domain.ShopOnboardingNudgeAnchor{},
		&domain.ShopOnboardingNudgeSent{},
	} {
		if !migrator.HasTable(model) {
			if err := r.db.AutoMigrate(model); err != nil {
				return err
			}
		}
	}

	// Seed the 4 templates once, if the table is empty — a fresh deploy
	// should have ready-to-send copy without a manual setup step.
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.OnboardingNudgeTemplate{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		if err := r.db.WithContext(ctx).Create(defaultTemplates()).Error; err != nil {
			return err
		}
	}

	// Seed the single settings row once.
	var settingsCount int64
	if err := r.db.WithContext(ctx).Model(&domain.OnboardingNudgeSettings{}).Count(&settingsCount).Error; err != nil {
		return err
	}
	if settingsCount == 0 {
		defaults := domain.OnboardingNudgeSettings{
			ID:           domain.OnboardingNudgeSettingsID,
			Enabled:      true,
			GapHours:     2,
			DurationDays: 7,
		}
		if err := r.db.WithContext(ctx).Create(&defaults).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *onboardingNudgeRepository) GetTemplates(ctx context.Context) ([]domain.OnboardingNudgeTemplate, error) {
	if err := r.ensureTables(ctx); err != nil {
		return nil, err
	}
	var templates []domain.OnboardingNudgeTemplate
	err := r.db.WithContext(ctx).Order("key").Find(&templates).Error
	return templates, err
}

func (r *onboardingNudgeRepository) SaveTemplate(ctx context.Context, tmpl domain.OnboardingNudgeTemplate) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Model(&domain.OnboardingNudgeTemplate{}).
		Where("key = ?", tmpl.Key).
		Updates(map[string]interface{}{
			"title":     tmpl.Title,
			"body":      tmpl.Body,
			"image_url": tmpl.ImageURL,
			"route":     tmpl.Route,
		}).Error
}

func (r *onboardingNudgeRepository) GetSettings(ctx context.Context) (domain.OnboardingNudgeSettings, error) {
	if err := r.ensureTables(ctx); err != nil {
		return domain.OnboardingNudgeSettings{}, err
	}
	var settings domain.OnboardingNudgeSettings
	err := r.db.WithContext(ctx).First(&settings, "id = ?", domain.OnboardingNudgeSettingsID).Error
	return settings, err
}

func (r *onboardingNudgeRepository) SaveSettings(ctx context.Context, settings domain.OnboardingNudgeSettings) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}
	settings.ID = domain.OnboardingNudgeSettingsID
	return r.db.WithContext(ctx).
		Model(&domain.OnboardingNudgeSettings{}).
		Where("id = ?", domain.OnboardingNudgeSettingsID).
		Updates(map[string]interface{}{
			"enabled":       settings.Enabled,
			"gap_hours":     settings.GapHours,
			"duration_days": settings.DurationDays,
		}).Error
}

func (r *onboardingNudgeRepository) RecordGoLiveOnce(ctx context.Context, shopID string, at time.Time) (bool, error) {
	if err := r.ensureTables(ctx); err != nil {
		return false, err
	}
	anchor := domain.ShopOnboardingNudgeAnchor{ShopID: shopID, GoLiveAt: at}
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&anchor)
	return result.RowsAffected > 0, result.Error
}

func (r *onboardingNudgeRepository) RecentlyLiveShops(ctx context.Context, since time.Time) ([]domain.OnboardingNudgeCandidate, error) {
	var shops []domain.OnboardingNudgeCandidate
	err := r.db.WithContext(ctx).
		Table("shop_details").
		Select("id AS shop_id, updated_at AS go_live_at").
		Where("shop_status = ? AND updated_at >= ?", "active", since).
		Scan(&shops).Error
	return shops, err
}

func (r *onboardingNudgeRepository) ActiveCandidates(ctx context.Context, now time.Time, durationDays int) ([]domain.OnboardingNudgeCandidate, error) {
	if err := r.ensureTables(ctx); err != nil {
		return nil, err
	}
	windowStart := now.AddDate(0, 0, -durationDays)
	var candidates []domain.OnboardingNudgeCandidate
	err := r.db.WithContext(ctx).
		Table("shop_onboarding_nudge_anchors a").
		Select("a.shop_id, a.go_live_at, sd.shop_name, sd.city").
		Joins("JOIN shop_details sd ON sd.id = a.shop_id").
		Where("a.go_live_at <= ? AND a.go_live_at >= ?", now, windowStart).
		Scan(&candidates).Error
	return candidates, err
}

func (r *onboardingNudgeRepository) SentSlotsByShop(ctx context.Context, shopIDs []string) (map[string]map[[2]int]bool, error) {
	result := make(map[string]map[[2]int]bool)
	if len(shopIDs) == 0 {
		return result, nil
	}
	var rows []domain.ShopOnboardingNudgeSent
	if err := r.db.WithContext(ctx).Where("shop_id IN ?", shopIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		if result[row.ShopID] == nil {
			result[row.ShopID] = make(map[[2]int]bool)
		}
		result[row.ShopID][[2]int{row.Day, row.Slot}] = true
	}
	return result, nil
}

func (r *onboardingNudgeRepository) MarkSent(ctx context.Context, shopID string, day, slot int) error {
	sent := domain.ShopOnboardingNudgeSent{ShopID: shopID, Day: day, Slot: slot}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&sent).Error
}

func (r *onboardingNudgeRepository) Stats(ctx context.Context, todayStart time.Time) (domain.OnboardingNudgeStats, error) {
	db := r.db.WithContext(ctx).Model(&domain.ShopOnboardingNudgeSent{})
	var stats domain.OnboardingNudgeStats

	if err := db.Session(&gorm.Session{}).Count(&stats.TotalSent).Error; err != nil {
		return stats, err
	}
	if err := db.Session(&gorm.Session{}).Where("sent_at >= ?", todayStart).Count(&stats.SentToday).Error; err != nil {
		return stats, err
	}
	if err := db.Session(&gorm.Session{}).Distinct("shop_id").Count(&stats.ShopsReached).Error; err != nil {
		return stats, err
	}

	var slotRows []struct {
		Slot int
		Sent int64
	}
	if err := db.Session(&gorm.Session{}).Select("slot, COUNT(*) AS sent").Group("slot").Scan(&slotRows).Error; err != nil {
		return stats, err
	}
	bySlot := make(map[int]int64, len(slotRows))
	for _, row := range slotRows {
		bySlot[row.Slot] = row.Sent
	}
	for i, key := range domain.NudgeTemplateOrder {
		stats.BySlot = append(stats.BySlot, domain.OnboardingNudgeSlotStat{Slot: i, Key: key, Sent: bySlot[i]})
	}

	from := todayStart.AddDate(0, 0, -6)
	var dayRows []struct {
		Day  time.Time
		Sent int64
	}
	if err := db.Session(&gorm.Session{}).
		Select("DATE(sent_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS sent").
		Where("sent_at >= ?", from).
		Group("day").Scan(&dayRows).Error; err != nil {
		return stats, err
	}
	byDay := make(map[string]int64, len(dayRows))
	for _, row := range dayRows {
		byDay[row.Day.Format("2006-01-02")] = row.Sent
	}
	for d := from; !d.After(todayStart); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		stats.Last7Days = append(stats.Last7Days, domain.OnboardingNudgeDayStat{Date: date, Sent: byDay[date]})
	}
	return stats, nil
}

func (r *onboardingNudgeRepository) SentCountsByShop(ctx context.Context, shopIDs []string) (map[string]domain.ShopOnboardingNudgeCount, error) {
	result := make(map[string]domain.ShopOnboardingNudgeCount)
	if len(shopIDs) == 0 {
		return result, nil
	}
	var rows []struct {
		ShopID string
		Slot   int
		Sent   int64
	}
	if err := r.db.WithContext(ctx).Model(&domain.ShopOnboardingNudgeSent{}).
		Select("shop_id, slot, COUNT(*) AS sent").
		Where("shop_id IN ?", shopIDs).
		Group("shop_id, slot").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.Slot < 0 || row.Slot >= len(domain.NudgeTemplateOrder) {
			continue
		}
		c := result[row.ShopID]
		if c.ByKey == nil {
			c.ByKey = make(map[string]int64)
		}
		c.ByKey[domain.NudgeTemplateOrder[row.Slot]] = row.Sent
		c.Total += row.Sent
		result[row.ShopID] = c
	}
	return result, nil
}
