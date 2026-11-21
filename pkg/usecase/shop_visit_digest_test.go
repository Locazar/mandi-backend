package usecase

import (
	"testing"
	"time"
)

func TestShopVisitDigestSchedule(t *testing.T) {
	at := func(h, m int) time.Time { return time.Date(2026, 11, 22, h, m, 0, 0, istZone) }
	if got := NextShopVisitDigestRun(at(10, 0)); !got.Equal(at(23, 50)) {
		t.Fatal(got)
	}
	if got := NextShopVisitDigestRun(at(23, 50)); !got.Equal(at(23, 50).AddDate(0, 0, 1)) {
		t.Fatal(got)
	}
	if got := NextShopVisitDigestRun(time.Date(2026, 11, 22, 18, 30, 0, 0, time.UTC)); !got.Equal(time.Date(2026, 11, 23, 23, 50, 0, 0, istZone)) {
		t.Fatal("utc", got)
	}
	if !ShouldRunShopVisitDigestNow(at(23, 55)) || ShouldRunShopVisitDigestNow(at(23, 49)) || ShouldRunShopVisitDigestNow(at(0, 5)) {
		t.Fatal("window")
	}
	if shopVisitDigestBody(1) != "1 customer visited your shop profile today. Add more products to reach more customers." {
		t.Fatal()
	}
	if shopVisitDigestBody(12) != "12 customers visited your shop profile today. Add more products to reach more customers." {
		t.Fatal()
	}
}
