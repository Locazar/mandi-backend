// Command onboarding-nudge-backfill anchors every currently-active shop into
// the onboarding-nudge sequence (see cmd/onboarding-nudges) as of "now".
//
// Why this exists: a shop only gets an anchor when it's approved (shop_status
// → active) — see adminUseCase.ApproveShop. Shops approved before the
// onboarding-nudge feature shipped never went through that code path, so
// without this they'd never receive the 4 reminders at all. Run this once
// after deploying the feature so every existing live shop starts its 7-day
// sequence today; idempotent (RecordGoLiveOnce is ON CONFLICT DO NOTHING),
// so it's safe to re-run — shops that already have an anchor are skipped.
//
// Usage (reads the same env/.env as the API):
//
//	go run ./cmd/onboarding-nudge-backfill            # dry-run: prints what would be anchored
//	go run ./cmd/onboarding-nudge-backfill -apply     # actually writes anchors
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/repository"
)

func main() {
	apply := flag.Bool("apply", false, "write anchors (default is dry-run)")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("load config: ", err)
	}

	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		log.Fatal("connect db: ", err)
	}

	var shopIDs []string
	if err := gormDB.Raw(`SELECT id FROM shop_details WHERE shop_status = 'active'`).Scan(&shopIDs).Error; err != nil {
		log.Fatal("query active shops: ", err)
	}
	log.Printf("found %d active shop(s)", len(shopIDs))

	nudgeRepo := repository.NewOnboardingNudgeRepository(gormDB)
	ctx := context.Background()
	now := time.Now()

	anchored, failed := 0, 0
	for _, id := range shopIDs {
		fmt.Printf("shop %s ← anchor at %s\n", id, now.Format(time.RFC3339))
		if !*apply {
			continue
		}
		if err := nudgeRepo.RecordGoLiveOnce(ctx, id, now); err != nil {
			log.Printf("WARN: anchor failed for %s: %v", id, err)
			failed++
			continue
		}
		anchored++
	}

	if *apply {
		log.Printf("done: %d anchored, %d failed (shops with an existing anchor are silently skipped by RecordGoLiveOnce)", anchored, failed)
	} else {
		log.Printf("dry-run complete (%d candidate shop(s)) — re-run with -apply to write", len(shopIDs))
	}
}
