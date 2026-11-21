package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/service/notification"
	uci "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"google.golang.org/api/iterator"
	"gorm.io/gorm"
)

// Daily shop-visit digest.
//
// Every night at 23:50 IST each seller whose shop profile was opened that day
// gets one push notification (no WhatsApp message):
//
//	"12 customers visited your shop profile today. Add more products to reach more customers."

//
// Visits come from the Firestore collections "shop_action_logs" (written by
// the customer web app) and "shop_action_log" (action = "profile_view");
// both are read and merged. One browser session counts once
// per shop per day, so refreshing the page does not inflate the number.
// A run is recorded per day in shop_visit_digest_runs so a restart never
// sends the same day's digest twice.

// IST is fixed at UTC+5:30 (India has no daylight saving), so a fixed zone
// avoids depending on tzdata being installed on the server.
var istZone = time.FixedZone("IST", 5*60*60+30*60)

const (
	shopVisitDigestHour   = 23
	shopVisitDigestMinute = 50
	shopActionProfileView = "profile_view"
)

// shopActionLogCollections are read and merged; a missing collection simply
// returns no documents.
var shopActionLogCollections = []string{"shop_action_logs", "shop_action_log"}

type ShopVisitDigestResult struct {
	Day         string
	AlreadySent bool
	Shops       int
	Sent        int
	Skipped     int
	Errors      int
}

type ShopVisitDigest struct {
	db           *gorm.DB
	notification uci.NotificationUseCase

	// fetchVisitors returns shop_id -> set of visitor session IDs for
	// [start, end). Replaceable in tests.
	fetchVisitors func(ctx context.Context, start, end time.Time) (map[string]map[string]struct{}, error)
}

func NewShopVisitDigest(db *gorm.DB, notificationUC uci.NotificationUseCase) *ShopVisitDigest {
	return &ShopVisitDigest{
		db:            db,
		notification:  notificationUC,
		fetchVisitors: fetchProfileVisitorsFromFirestore,
	}
}

// shopVisitDigestBody is the notification text sent to the seller.
func shopVisitDigestBody(visitors int) string {
	if visitors == 1 {
		return "1 customer visited your shop profile today. Add more products to reach more customers."
	}
	return fmt.Sprintf("%d customers visited your shop profile today. Add more products to reach more customers.", visitors)
}

// NextShopVisitDigestRun returns the next 23:50 IST strictly after now.
func NextShopVisitDigestRun(now time.Time) time.Time {
	n := now.In(istZone)
	run := time.Date(n.Year(), n.Month(), n.Day(), shopVisitDigestHour, shopVisitDigestMinute, 0, 0, istZone)
	if !run.After(n) {
		run = run.AddDate(0, 0, 1)
	}
	return run
}

// ShouldRunShopVisitDigestNow reports whether now falls in the daily window
// 23:50–23:59 IST, used on startup to catch up a run missed by a restart.
func ShouldRunShopVisitDigestNow(now time.Time) bool {
	n := now.In(istZone)
	return n.Hour() == shopVisitDigestHour && n.Minute() >= shopVisitDigestMinute
}

func (d *ShopVisitDigest) ensureTable(ctx context.Context) error {
	return d.db.WithContext(ctx).Exec(`CREATE TABLE IF NOT EXISTS shop_visit_digest_runs (
		day        date PRIMARY KEY,
		shops      integer NOT NULL DEFAULT 0,
		sent       integer NOT NULL DEFAULT 0,
		created_at timestamptz NOT NULL DEFAULT now()
	)`).Error
}

// RunForDay sends the digest for the IST calendar day containing `at`.
func (d *ShopVisitDigest) RunForDay(ctx context.Context, at time.Time) (ShopVisitDigestResult, error) {
	n := at.In(istZone)
	start := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, istZone)
	end := start.AddDate(0, 0, 1)
	day := start.Format("2006-01-02")
	res := ShopVisitDigestResult{Day: day}

	if err := d.ensureTable(ctx); err != nil {
		return res, fmt.Errorf("shop visit digest: table: %w", err)
	}
	var already int64
	if err := d.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM shop_visit_digest_runs WHERE day = ?`, day).Scan(&already).Error; err != nil {
		return res, fmt.Errorf("shop visit digest: check run: %w", err)
	}
	if already > 0 {
		res.AlreadySent = true
		return res, nil
	}

	visitors, err := d.fetchVisitors(ctx, start, end)
	if err != nil {
		return res, fmt.Errorf("shop visit digest: read visits: %w", err)
	}
	res.Shops = len(visitors)

	if len(visitors) > 0 {
		ids := make([]string, 0, len(visitors))
		for id := range visitors {
			ids = append(ids, id)
		}
		var owners []struct {
			ID      string `gorm:"column:id"`
			AdminID string `gorm:"column:admin_id"`
		}
		if err := d.db.WithContext(ctx).Table("shop_details").Select("id, admin_id").Where("id IN ?", ids).Scan(&owners).Error; err != nil {
			return res, fmt.Errorf("shop visit digest: load shops: %w", err)
		}

		for _, o := range owners {
			count := len(visitors[o.ID])
			if o.AdminID == "" || count == 0 {
				res.Skipped++
				continue
			}
			_, sendErr := d.notification.SendPushNotification(ctx, request.SendPushRequest{
				OwnerID:   o.AdminID,
				OwnerType: "seller",
				Title:     "Today's shop visits",
				Body:      shopVisitDigestBody(count),
				EventType: "shop_visit_digest",
				Data: map[string]string{
					"event_type": "shop_visit_digest",
					"shop_id":    o.ID,
					"visitors":   fmt.Sprint(count),
					"day":        day,
				},
			})
			if sendErr != nil {
				log.Printf("WARN [shop-visit-digest]: send shop=%s failed: %v", o.ID, sendErr)
				res.Errors++
				continue
			}
			res.Sent++
		}
	}

	if err := d.db.WithContext(ctx).Exec(
		`INSERT INTO shop_visit_digest_runs (day, shops, sent) VALUES (?, ?, ?) ON CONFLICT (day) DO NOTHING`,
		day, res.Shops, res.Sent).Error; err != nil {
		log.Printf("WARN [shop-visit-digest]: record run %s failed: %v", day, err)
	}
	return res, nil
}

// fetchProfileVisitorsFromFirestore reads one day of shop action logs.
// Only the created_at range is queried (single-field index, no composite
// index needed); the action filter is applied here.
func fetchProfileVisitorsFromFirestore(ctx context.Context, start, end time.Time) (map[string]map[string]struct{}, error) {
	client, err := notification.SharedFirestoreClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestore client: %w", err)
	}
	out := map[string]map[string]struct{}{}
	for _, coll := range shopActionLogCollections {
		if err := readProfileViews(ctx, client.Collection(coll).
			Where("created_at", ">=", start).
			Where("created_at", "<", end).
			Documents(ctx), out); err != nil {
			return nil, fmt.Errorf("%s: %w", coll, err)
		}
	}
	return out, nil
}

func readProfileViews(ctx context.Context, iter *firestore.DocumentIterator, out map[string]map[string]struct{}) error {
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return err
		}
		data := doc.Data()
		if action, _ := data["action"].(string); action != shopActionProfileView {
			continue
		}
		shopID := fmt.Sprint(data["shop_id"]) // tolerate numeric IDs
		if shopID == "" || shopID == "<nil>" {
			continue
		}
		visitor, _ := data["session_id"].(string)
		if visitor == "" {
			visitor = doc.Ref.ID // no session: count the view on its own
		}
		if out[shopID] == nil {
			out[shopID] = map[string]struct{}{}
		}
		out[shopID][visitor] = struct{}{}
	}
}
