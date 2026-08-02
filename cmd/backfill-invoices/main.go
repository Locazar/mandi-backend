// Command backfill-invoices issues tax invoices for subscription orders that
// were paid before invoicing existed.
//
// Idempotent: orders that already have an invoice are skipped, so re-running is
// safe. Orders are processed oldest-paid-first so backfilled numbers run
// chronologically.
//
// RUN THIS BEFORE the invoicing build accepts any new payment. If a live
// payment issues a number first, backfilled and live invoices interleave in the
// sequence and the chronology is lost.
//
// Known limitation: shop name and address are snapshotted as of the backfill
// run, not as of the original payment — that history does not exist in the data.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository"
	"github.com/rohit221990/mandi-backend/pkg/service/crypto"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "report what would be issued without writing")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	ctx := context.Background()
	subRepo := repository.NewSubscriptionRepository(gormDB)
	invoiceRepo := repository.NewInvoiceRepository(gormDB)

	keys, err := crypto.ParseKeyring(cfg.PIIEncryptionKeys)
	if err != nil {
		log.Fatalf("parse PII encryption keyring: %v", err)
	}
	encSvc, err := crypto.NewService(keys, cfg.PIIEncryptionActiveKey)
	if err != nil {
		log.Fatalf("init crypto service: %v", err)
	}

	adminRepo := repository.NewAdminRepository(gormDB, encSvc)

	profile, err := invoiceRepo.GetCompanyBillingProfile(ctx)
	if err != nil {
		log.Fatalf("company billing profile missing — seed it before backfilling: %v", err)
	}
	if profile.LegalName == "" {
		log.Fatal("company billing profile has no legal name — configure it in the admin portal first")
	}

	orders, err := subRepo.FindAllPaidOrdersWithoutInvoice(ctx)
	if err != nil {
		log.Fatalf("find paid orders: %v", err)
	}

	fmt.Printf("Found %d paid order(s) without an invoice.\n", len(orders))
	if len(orders) == 0 {
		return
	}

	issued, failed := 0, 0
	for _, order := range orders {
		plan, err := subRepo.FindSubscriptionPlanByID(ctx, order.PlanID)
		if err != nil {
			log.Printf("SKIP order=%s: plan lookup failed: %v", order.ID, err)
			failed++
			continue
		}

		// GetShopByOwnerID decrypts document_value; a raw join would yield
		// ciphertext in the GSTIN field.
		shop, err := adminRepo.GetShopByOwnerID(ctx, order.UserID)
		if err != nil {
			log.Printf("WARN order=%s: shop lookup failed, issuing without buyer details: %v", order.ID, err)
			shop = domain.ShopDetails{}
		}

		paidAt := order.PaidAt
		if paidAt == nil {
			log.Printf("SKIP order=%s: paid order has no paid_at", order.ID)
			failed++
			continue
		}

		if *dryRun {
			fmt.Printf("WOULD ISSUE order=%s fy=%s paid_at=%s amount=%s\n",
				order.ID, domain.FinancialYear(*paidAt), paidAt.Format("2006-01-02"), order.Price.String())
			issued++
			continue
		}

		// Allocate and insert in one transaction: a failed insert (a real DB
		// error, not the duplicate case this loop can't reach since the query
		// already excludes orders with an invoice) would otherwise burn the
		// allocated number permanently in a series required to be gapless.
		inv, err := invoiceRepo.CreateInvoiceWithSequence(ctx, domain.FinancialYear(*paidAt), func(seq int) domain.Invoice {
			return usecase.BuildInvoice(usecase.BuildInvoiceInput{
				Order:          order.SubscriptionOrder,
				Plan:           plan,
				Shop:           shop,
				Profile:        profile,
				SequenceNumber: seq,
			})
		})
		if err != nil {
			log.Printf("FAIL order=%s: create invoice: %v", order.ID, err)
			failed++
			continue
		}

		fmt.Printf("ISSUED %s for order=%s\n", inv.InvoiceNumber, order.ID)
		issued++
	}

	fmt.Printf("\nDone. issued=%d failed=%d dry_run=%v\n", issued, failed, *dryRun)
	if failed > 0 {
		os.Exit(1)
	}
}
