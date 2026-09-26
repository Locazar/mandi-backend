package usecase

import (
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// maxBillPaise is a sanity ceiling on a self-reported in-store bill (₹10,00,000).
// Points are capped anyway; this just rejects obvious typos.
const maxBillPaise int64 = 100_000_000

// rewardTZ is the timezone whose calendar day the daily caps reset on.
var rewardTZ = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return time.FixedZone("IST", 5*60*60+30*60)
	}
	return loc
}()

// CalculatePoints applies a reward formula to the net amount actually paid.
func CalculatePoints(f domain.RewardFormula, netPaise int64) int64 {
	if netPaise <= 0 {
		return 0
	}
	pts := f.BasePoints + (netPaise/10000)*f.PointsPer100Rupees
	if f.MaxPoints > 0 && pts > f.MaxPoints {
		pts = f.MaxPoints
	}
	if pts < 0 {
		return 0
	}
	return pts
}

// discountPaise is the store discount on a bill, rounded down to the paisa.
func discountPaise(billPaise int64, percent int) int64 {
	return billPaise * int64(percent) / 100
}

// distanceMeters is the haversine great-circle distance between two points.
func distanceMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusM = 6371000.0
	rad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat := rad(lat2 - lat1)
	dLng := rad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rad(lat1))*math.Cos(rad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earthRadiusM * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// rewardDayStart returns midnight (Asia/Kolkata) of the day containing t.
func rewardDayStart(t time.Time) time.Time {
	local := t.In(rewardTZ)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, rewardTZ)
}

// normalizePhone reduces a phone number to its last 10 digits, or "" when it
// has fewer than 10 — so "+91 98765 43210" and "09876543210" compare equal.
func normalizePhone(s string) string {
	digits := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, s)
	if len(digits) < 10 {
		return ""
	}
	return digits[len(digits)-10:]
}
