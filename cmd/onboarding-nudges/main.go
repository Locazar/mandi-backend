// Command onboarding-nudges sends the 4 automated onboarding reminders (add
// products, update shop photo, update current address, view your shop on
// locazar.in) to every shop still inside its 7-day post-approval window,
// spaced by the configured gap (default 2 hours), for the configured
// duration (default 7 days) — both editable from admin-portal.
//
// Meant to run on a schedule (Cloud Run Job + Cloud Scheduler, same pattern
// as cmd/enquiry-autoreject), not as an in-process ticker in the API server.
// Every (shop, day, slot) that's due and not yet sent gets sent on each
// invocation — safe to run on any cadence (e.g. every 15-30 minutes): a
// missed or delayed run just catches up on the next one, and the sent
// ledger (shop_onboarding_nudge_sents) guarantees nothing is ever sent
// twice.
//
// Usage (reads the same env/.env as the API):
//
//	go run ./cmd/onboarding-nudges
package main

import (
	"context"
	"log"

	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/di"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("load config: ", err)
	}

	uc, err := di.InitializeOnboardingNudgeUseCase(cfg)
	if err != nil {
		log.Fatal("initialize onboarding nudge use case: ", err)
	}

	result, err := uc.RunSweep(context.Background())
	if err != nil {
		log.Fatal("sweep failed: ", err)
	}
	if result.Disabled {
		log.Printf("onboarding-nudges: disabled in settings — nothing sent")
		return
	}
	log.Printf("onboarding-nudges: sent=%d skipped=%d errors=%d", result.Sent, result.Skipped, result.Errors)
}
