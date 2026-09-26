package usecase

import (
	"math"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestCalculatePoints(t *testing.T) {
	seller := domain.RewardFormula{BasePoints: 10, PointsPer100Rupees: 1, MaxPoints: 50}
	tests := map[string]struct {
		f    domain.RewardFormula
		net  int64
		want int64
	}{
		"zero net earns nothing":          {seller, 0, 0},
		"negative net earns nothing":      {seller, -500, 0},
		"under 100 rupees is base only":   {seller, 9999, 10},
		"250 rupees adds two":             {seller, 25000, 12},
		"capped at max":                   {seller, 1_000_000, 50},
		"max zero means uncapped":         {domain.RewardFormula{BasePoints: 10, PointsPer100Rupees: 1}, 1_000_000, 110},
		"customer formula smaller values": {domain.RewardFormula{BasePoints: 5, PointsPer100Rupees: 1, MaxPoints: 25}, 45000, 9},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, CalculatePoints(tc.f, tc.net))
		})
	}
}

func TestDiscountPaiseFloors(t *testing.T) {
	assert.Equal(t, int64(9999), discountPaise(99999, 10))
	assert.Equal(t, int64(0), discountPaise(50000, 0))
	assert.Equal(t, int64(10000), discountPaise(50000, 20))
}

func TestDistanceMeters(t *testing.T) {
	assert.InDelta(t, 0, distanceMeters(12.9756, 77.6050, 12.9756, 77.6050), 0.001)
	// 0.0018° of latitude ≈ 200 m.
	d := distanceMeters(12.9756, 77.6050, 12.9774, 77.6050)
	assert.InDelta(t, 200.2, d, 1.0)
	assert.False(t, math.IsNaN(distanceMeters(0, 0, 0, 0)))
}

func TestRewardDayStartUsesIST(t *testing.T) {
	// 2026-09-26 20:00 UTC is 2026-09-27 01:30 IST → IST day starts 2026-09-26 18:30 UTC.
	got := rewardDayStart(time.Date(2026, 9, 26, 20, 0, 0, 0, time.UTC))
	assert.True(t, got.Equal(time.Date(2026, 9, 26, 18, 30, 0, 0, time.UTC)), got.UTC().String())

	// 2026-09-26 17:00 UTC is 22:30 IST on the 26th → day started 2026-09-25 18:30 UTC.
	got = rewardDayStart(time.Date(2026, 9, 26, 17, 0, 0, 0, time.UTC))
	assert.True(t, got.Equal(time.Date(2026, 9, 25, 18, 30, 0, 0, time.UTC)), got.UTC().String())
}

func TestNormalizePhone(t *testing.T) {
	assert.Equal(t, "9876543210", normalizePhone("+91 98765 43210"))
	assert.Equal(t, "9876543210", normalizePhone("09876543210"))
	assert.Equal(t, "9876543210", normalizePhone("9876543210"))
	assert.Equal(t, "", normalizePhone(""))
	assert.Equal(t, "", normalizePhone("12345"))
}
