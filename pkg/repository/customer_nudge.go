package repository

import (
	"context"
	"sync"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type customerNudgeRepository struct {
	db *gorm.DB
}

func NewCustomerNudgeRepository(db *gorm.DB) interfaces.CustomerNudgeRepository {
	return &customerNudgeRepository{db: db}
}

func defaultCustomerTemplates() []domain.CustomerNudgeTemplate {
	return []domain.CustomerNudgeTemplate{
		{Key: domain.CustomerNudgeExploreShops, Title: "Shops near you are waiting 🛍️", Body: "Hi {{name}}, discover verified local shops and their latest products around you on Locazar."},
		{Key: domain.CustomerNudgeSetLocation, Title: "Set your location 📍", Body: "Add your address so Locazar can show you the closest shops and the best local deals."},
		{Key: domain.CustomerNudgeSendEnquiry, Title: "Ask a shop directly 💬", Body: "Found something you like? Send an enquiry and negotiate the price straight with the shop."},
		{Key: domain.CustomerNudgeShareApp, Title: "Love local? Share Locazar ❤️", Body: "Invite friends and family to discover the best shops in your neighbourhood."},
	}
}

// ensureMu serialises the lazy migrate-and-seed: admin-portal loads templates,
// settings and stats in parallel on first open, and concurrent runs would race
// on CREATE TABLE and the seed inserts.
var ensureMu sync.Mutex

func (r *customerNudgeRepository) ensureTables(ctx context.Context) error {
	ensureMu.Lock()
	defer ensureMu.Unlock()
	migrator := r.db.Migrator()
	for _, model := range []interface{}{
		&domain.CustomerNudgeTemplate{},
		&domain.CustomerNudgeSettings{},
		&domain.CustomerNudgeSent{},
	} {
		if !migrator.HasTable(model) {
			if err := r.db.AutoMigrate(model); err != nil {
				return err
			}
		}
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.CustomerNudgeTemplate{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		if err := r.db.WithContext(ctx).Create(defaultCustomerTemplates()).Error; err != nil {
			return err
		}
	}

	var settingsCount int64
	if err := r.db.WithContext(ctx).Model(&domain.CustomerNudgeSettings{}).Count(&settingsCount).Error; err != nil {
		return err
	}
	if settingsCount == 0 {
		// Off by default: turning it on is a deliberate admin decision, since
		// it messages real customers.
		defaults := domain.CustomerNudgeSettings{ID: domain.OnboardingNudgeSettingsID, Enabled: false, GapHours: 2, DurationDays: 7, StartsAt: time.Now()}
		return r.db.WithContext(ctx).Create(&defaults).Error
	}
	return nil
}

func (r *customerNudgeRepository) GetTemplates(ctx context.Context) ([]domain.CustomerNudgeTemplate, error) {
	if err := r.ensureTables(ctx); err != nil {
		return nil, err
	}
	var templates []domain.CustomerNudgeTemplate
	err := r.db.WithContext(ctx).Order("key").Find(&templates).Error
	return templates, err
}

func (r *customerNudgeRepository) SaveTemplate(ctx context.Context, tmpl domain.CustomerNudgeTemplate) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&domain.CustomerNudgeTemplate{}).Where("key = ?", tmpl.Key).
		Updates(map[string]interface{}{"title": tmpl.Title, "body": tmpl.Body, "image_url": tmpl.ImageURL, "route": tmpl.Route}).Error
}

func (r *customerNudgeRepository) GetSettings(ctx context.Context) (domain.CustomerNudgeSettings, error) {
	if err := r.ensureTables(ctx); err != nil {
		return domain.CustomerNudgeSettings{}, err
	}
	var settings domain.CustomerNudgeSettings
	err := r.db.WithContext(ctx).First(&settings, "id = ?", domain.OnboardingNudgeSettingsID).Error
	return settings, err
}

func (r *customerNudgeRepository) SaveSettings(ctx context.Context, settings domain.CustomerNudgeSettings) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&domain.CustomerNudgeSettings{}).Where("id = ?", domain.OnboardingNudgeSettingsID).
		Updates(map[string]interface{}{
			"enabled": settings.Enabled, "gap_hours": settings.GapHours,
			"duration_days": settings.DurationDays, "starts_at": settings.StartsAt,
		}).Error
}

func (r *customerNudgeRepository) Candidates(ctx context.Context, since, now time.Time) ([]domain.CustomerNudgeCandidate, error) {
	var out []domain.CustomerNudgeCandidate
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id AS user_id, first_name, created_at AS start_at").
		Where("deleted_at IS NULL AND block_status = ? AND created_at >= ? AND created_at <= ?", false, since, now).
		Scan(&out).Error
	return out, err
}

func (r *customerNudgeRepository) SentSlotsByUser(ctx context.Context, userIDs []string) (map[string]map[[2]int]bool, error) {
	result := make(map[string]map[[2]int]bool)
	if len(userIDs) == 0 {
		return result, nil
	}
	var rows []domain.CustomerNudgeSent
	if err := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		if result[row.UserID] == nil {
			result[row.UserID] = make(map[[2]int]bool)
		}
		result[row.UserID][[2]int{row.Day, row.Slot}] = true
	}
	return result, nil
}

func (r *customerNudgeRepository) MarkSent(ctx context.Context, userID string, day, slot int, delivered bool) error {
	row := domain.CustomerNudgeSent{UserID: userID, Day: day, Slot: slot, Delivered: delivered}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

func (r *customerNudgeRepository) Stats(ctx context.Context, todayStart time.Time) (domain.CustomerNudgeStats, error) {
	if err := r.ensureTables(ctx); err != nil {
		return domain.CustomerNudgeStats{}, err
	}
	base := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&domain.CustomerNudgeSent{}).Where("delivered = ?", true)
	}
	var stats domain.CustomerNudgeStats
	if err := base().Count(&stats.TotalSent).Error; err != nil {
		return stats, err
	}
	if err := base().Where("sent_at >= ?", todayStart).Count(&stats.SentToday).Error; err != nil {
		return stats, err
	}
	if err := base().Distinct("user_id").Count(&stats.CustomersReached).Error; err != nil {
		return stats, err
	}
	if err := r.db.WithContext(ctx).Model(&domain.CustomerNudgeSent{}).Where("delivered = ?", false).Count(&stats.NotDelivered).Error; err != nil {
		return stats, err
	}

	var slotRows []struct {
		Slot int
		Sent int64
	}
	if err := base().Select("slot, COUNT(*) AS sent").Group("slot").Scan(&slotRows).Error; err != nil {
		return stats, err
	}
	bySlot := make(map[int]int64, len(slotRows))
	for _, row := range slotRows {
		bySlot[row.Slot] = row.Sent
	}
	for i, key := range domain.CustomerNudgeTemplateOrder {
		stats.BySlot = append(stats.BySlot, domain.OnboardingNudgeSlotStat{Slot: i, Key: key, Sent: bySlot[i]})
	}

	from := todayStart.AddDate(0, 0, -6)
	// Aliased sent_date, not day: the ledger has a "day" column, and GROUP BY
	// would bind to that column instead of the alias.
	var dayRows []struct {
		SentDate time.Time
		Sent     int64
	}
	if err := base().Select("DATE(sent_at AT TIME ZONE 'UTC') AS sent_date, COUNT(*) AS sent").
		Where("sent_at >= ?", from).Group("sent_date").Scan(&dayRows).Error; err != nil {
		return stats, err
	}
	byDay := make(map[string]int64, len(dayRows))
	for _, row := range dayRows {
		byDay[row.SentDate.Format("2006-01-02")] = row.Sent
	}
	for d := from; !d.After(todayStart); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		stats.Last7Days = append(stats.Last7Days, domain.OnboardingNudgeDayStat{Date: date, Sent: byDay[date]})
	}
	return stats, nil
}
