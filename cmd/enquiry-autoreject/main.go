// Command enquiry-autoreject sweeps the Firestore "enquiry" collection for
// negotiations that have sat too long awaiting a reply from whichever side
// owes one, and auto-rejects them with a reason naming the non-responsive
// party. Meant to be run on a schedule (Cloud Run Job + Cloud Scheduler),
// not embedded as an in-process ticker in the API server — the API scales
// across replicas and to zero, so a ticker there would double-fire or never
// fire.
//
// Whose turn it is to reply is encoded in the enquiry's status string
// itself (no separate "turn" field) — the same classification already used
// by enquiryRecipientResolver in pkg/service/notification/watcher_rules.go
// to route live notifications:
//
//	seller owes a reply:   new, pending_seller_price, pending_seller_final, seller_final_update
//	customer owes a reply: pending_customer_price, pending_customer_final
//	turn ambiguous:        in_progress, counter_offer
//
// The ambiguous-turn bucket exists so NO enquiry survives past the window in
// any non-terminal status. Sensitive states are deliberately excluded and
// never auto-rejected: customer_accepted_final (an agreed deal), on_hold (an
// intentional pause), and dispute/dispute_resolved (need human resolution).
//
// "Too long" is measured from Firestore's own server-assigned document
// UpdateTime (not an app-set field — no such field is reliably populated
// across every enquiry-writing code path), compared against
// ENQUIRY_AUTO_REJECT_HOURS (config.EnquiryAutoRejectHours). A value of 0
// disables the sweep entirely — ships dark by default.
//
// Usage (reads the same env/.env as the API):
//
//	go run ./cmd/enquiry-autoreject            # dry-run: prints what would be rejected
//	go run ./cmd/enquiry-autoreject -apply     # actually rejects + notifies
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/di"
	notificationSvc "github.com/rohit221990/mandi-backend/pkg/service/notification"
	"google.golang.org/api/iterator"
)

const enquiryCollection = "enquiry"

var sellerOwedStatuses = []string{
	"new",
	"pending_seller_price",
	"pending_seller_final",
	"seller_final_update",
}

var customerOwedStatuses = []string{
	"pending_customer_price",
	"pending_customer_final",
}

// inactiveStatuses are active negotiation states where a reply is expected
// but the turn-owner is ambiguous — swept so nothing lingers past the window,
// rejected with a generic reason and both parties notified.
var inactiveStatuses = []string{
	"in_progress",
	"counter_offer",
}

func main() {
	apply := flag.Bool("apply", false, "reject stale enquiries and send notifications (default is dry-run)")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("load config: ", err)
	}

	if cfg.EnquiryAutoRejectHours <= 0 {
		log.Printf("ENQUIRY_AUTO_REJECT_HOURS is 0/unset — sweep disabled, exiting")
		return
	}
	cutoff := time.Now().Add(-time.Duration(cfg.EnquiryAutoRejectHours) * time.Hour)

	notificationSvc.InitSharedFirebaseApp(cfg.FirebaseProjectID, cfg.FirebaseConfig)
	ctx := context.Background()
	fs, err := notificationSvc.SharedFirestoreClient(ctx)
	if err != nil {
		log.Fatal("firestore client: ", err)
	}
	defer fs.Close()

	notificationUC, err := di.InitializeNotificationUseCase(cfg)
	if err != nil {
		log.Fatal("initialize notification use case: ", err)
	}

	checked, rejected, notifyFailures := 0, 0, 0

	sweep := func(statuses []string, reason string, rejectedBy string, side string) {
		iter := fs.Collection(enquiryCollection).Where("status", "in", statuses).Documents(ctx)
		defer iter.Stop()
		for {
			docSnap, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("WARN: %s sweep: iterate: %v", side, err)
				break
			}
			checked++

			if docSnap.UpdateTime.After(cutoff) {
				continue // still within the grace period
			}

			data := docSnap.Data()
			userID := docString(data, "userId", "user_id")
			sellerID := docString(data, "sellerId", "seller_id", "shopId", "shop_id")

			short := docSnap.Ref.ID
			fmt.Printf("%s: %s (%s, last updated %s) → reject: %s\n",
				side, short, docSnap.Ref.Path, docSnap.UpdateTime.Format(time.RFC3339), reason)

			if !*apply {
				continue
			}

			_, err = docSnap.Ref.Update(ctx, []firestore.Update{
				{Path: "status", Value: "completed_rejected"},
				{Path: "finalStatus", Value: "rejected"},
				{Path: "rejectedBy", Value: rejectedBy},
				{Path: "autoRejectReason", Value: reason},
				{Path: "autoRejectedAt", Value: firestore.ServerTimestamp},
			})
			if err != nil {
				log.Printf("WARN: reject failed for %s: %v", docSnap.Ref.Path, err)
				continue
			}
			rejected++

			if userID != "" {
				if err := notificationUC.SendPushNotification(ctx, request.SendPushRequest{
					OwnerID:   userID,
					OwnerType: "user",
					Title:     "Enquiry auto-rejected",
					Body:      reason,
					EventType: "enquiry_auto_rejected",
					Data:      map[string]string{"enquiry_id": docSnap.Ref.ID},
				}); err != nil {
					notifyFailures++
					log.Printf("WARN: notify user %s failed: %v", userID, err)
				}
			}
			if sellerID != "" {
				if err := notificationUC.SendPushNotification(ctx, request.SendPushRequest{
					OwnerID:   sellerID,
					OwnerType: "seller",
					Title:     "Enquiry auto-rejected",
					Body:      reason,
					EventType: "enquiry_auto_rejected",
					Data:      map[string]string{"enquiry_id": docSnap.Ref.ID},
				}); err != nil {
					notifyFailures++
					log.Printf("WARN: notify seller %s failed: %v", sellerID, err)
				}
			}
		}
	}

	sweep(sellerOwedStatuses, "Did not get reply from seller", "system_seller_timeout", "seller-owed")
	sweep(customerOwedStatuses, "Did not get reply from customer", "system_customer_timeout", "customer-owed")
	sweep(inactiveStatuses, "Enquiry closed due to inactivity", "system_inactivity_timeout", "inactive")

	if *apply {
		log.Printf("done: checked %d, rejected %d, notify failures %d", checked, rejected, notifyFailures)
	} else {
		log.Printf("dry-run complete: checked %d, %d would be rejected — re-run with -apply to write", checked, rejected)
	}
}

// docString reads the first present, non-empty string field among keys —
// enquiry documents have used both camelCase and snake_case field names
// across app versions.
func docString(doc map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := doc[key]; ok && value != nil {
			text := strings.TrimSpace(fmt.Sprintf("%v", value))
			if text != "" && text != "<nil>" {
				return text
			}
		}
	}
	return ""
}
