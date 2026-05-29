package repository

import (
	"context"
	"math"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type bannerRepository struct {
	db *gorm.DB
}

func NewBannerRepository(db *gorm.DB) repo.BannerRepository {
	return &bannerRepository{
		db: db,
	}
}

func (r *bannerRepository) GetActiveBanners(ctx context.Context, departmentID, categoryID *string) ([]domain.Banner, error) {
	var banners []domain.Banner
	query := r.db.WithContext(ctx).Where("active = ?", true)

	if departmentID != nil && *departmentID != "" {
		query = query.Where("department_id = ?", *departmentID)
	}
	if categoryID != nil && *categoryID != "" {
		query = query.Where("category_id = ?", *categoryID)
	}

	err := query.Find(&banners).Error
	return banners, err
}

// GetBannersForUser fetches all active banners and filters in Go:
//  1. Geo     — banner center within radius, OR banner has no coords set (0,0) = global
//  2. Pincode — banner pincode must be empty OR match request pincode
//  3. Dept    — banner department_id must be nil OR match request department_id
//  4. Cat     — banner category_id must be nil OR match request category_id
//  5. AppType — banner app_type must be "" (global) OR "all" OR exactly match request app_type
func (r *bannerRepository) GetBannersForUser(ctx context.Context, req request.BannerFilterRequest) ([]domain.Banner, error) {
	var all []domain.Banner
	if err := r.db.WithContext(ctx).Where("active = ?", true).Order("created_at DESC").Find(&all).Error; err != nil {
		return nil, err
	}

	var result []domain.Banner
	for _, b := range all {

		// ── 1. Geo filter ──────────────────────────────────────────────────────
		// Skip geo check only if banner has no coordinates (global banner)
		if b.Latitude != 0 || b.Longitude != 0 {
			dist := haversineKm(req.Latitude, req.Longitude, b.Latitude, b.Longitude)
			allowedRadius := b.RadiusKm
			if allowedRadius <= 0 {
				allowedRadius = req.Radius
			}
			if dist > allowedRadius {
				continue
			}
		}

		// ── 2. Pincode filter ──────────────────────────────────────────────────
		// Skip banner only if BOTH sides have a pincode and they don't match
		if b.Pincode != "" && req.Pincode != "" && b.Pincode != req.Pincode {
			continue
		}

		// ── 3. Department filter ───────────────────────────────────────────────
		// nil department_id on banner = applies to all departments
		if req.DepartmentID > 0 && b.DepartmentID != nil && *b.DepartmentID != req.DepartmentID {
			continue
		}

		// ── 4. Category filter ─────────────────────────────────────────────────
		// nil category_id on banner = applies to all categories
		if req.CategoryID > 0 && b.CategoryID != nil && *b.CategoryID != req.CategoryID {
			continue
		}

		// ── 5. AppType filter ──────────────────────────────────────────────────
		// If request provides an app_type, only include:
		//   - banners with empty app_type (legacy/global banners)
		//   - banners with app_type = "all"
		//   - banners whose app_type exactly matches the request
		if req.AppType != "" {
			if b.AppType != "" && b.AppType != "all" && b.AppType != domain.BannerAppType(req.AppType) {
				continue
			}
		}

		result = append(result, b)
	}

	return result, nil
}

// haversineKm returns the great-circle distance in km between two lat/lng points.
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const earthR = 6371.0
	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return earthR * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func toRad(deg float64) float64 {
	return deg * math.Pi / 180
}