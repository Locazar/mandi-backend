// Package enquiryautoreject implements a Google Cloud Function (Gen 2, HTTP)
// that auto-rejects stale enquiries in the Firestore "enquiry" collection.
//
// It is the scheduled counterpart to the enquiry-notification-handler function:
// where that one reacts to live document changes, this one runs on a cadence
// (Cloud Scheduler → HTTP, hourly) and closes out enquiries that have sat too
// long without a reply, so nothing lingers in a non-terminal status past the
// configured window.
//
// Registered entry-point:
//   - AutoRejectStaleEnquiries — HTTP (invoked by Cloud Scheduler with OIDC)
//
// Whose turn it is to reply is encoded in the enquiry's status string itself
// (there is no separate "turn" field) — the same classification the live
// watcher/handler uses to route notifications:
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
// UpdateTime, compared against ENQUIRY_AUTO_REJECT_HOURS. A value of 0 (or
// unset) disables the sweep — the function returns 200 and does nothing.
//
// Config (all via environment; NO database or Firebase secret needed — the
// runtime service account's Application Default Credentials cover Firestore
// and FCM):
//   - GOOGLE_CLOUD_PROJECT / GCP_PROJECT : project id (auto-set on Cloud Run)
//   - ENQUIRY_AUTO_REJECT_HOURS          : window in hours (0/unset = disabled)
//
// A `?dryRun=true` query param lists what would be rejected without writing
// or notifying — safe to curl against the deployed function for verification.
package enquiryautoreject

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/rohit221990/mandi-backend/pkg/service/notification"
	"google.golang.org/api/iterator"
)

const enquiryCollection = "enquiry"

// autoRejectStatus is written to BOTH status and finalStatus when the sweep
// closes a stale enquiry, so a system timeout reads distinctly from a normal
// person-driven rejection (completed_rejected / rejected).
const autoRejectStatus = "Rejected_by_System"

// statusBucket is one class of stale enquiry and how to close it out.
type statusBucket struct {
	statuses   []string
	reason     string
	rejectedBy string
	label      string
}

var buckets = []statusBucket{
	{
		statuses:   []string{"new", "pending_seller_price", "pending_seller_final", "seller_final_update"},
		reason:     "Did not get reply from seller",
		rejectedBy: "system_seller_timeout",
		label:      "seller-owed",
	},
	{
		statuses:   []string{"pending_customer_price", "pending_customer_final"},
		reason:     "Did not get reply from customer",
		rejectedBy: "system_customer_timeout",
		label:      "customer-owed",
	},
	{
		// Ambiguous turn — swept so nothing lingers past the window; both
		// parties are notified. Sensitive states (customer_accepted_final,
		// on_hold, dispute, dispute_resolved) are intentionally NOT listed.
		statuses:   []string{"in_progress", "counter_offer"},
		reason:     "Enquiry closed due to inactivity",
		rejectedBy: "system_inactivity_timeout",
		label:      "inactive",
	},
}

func init() {
	functions.HTTP("AutoRejectStaleEnquiries", AutoRejectStaleEnquiries)
}

type runSummary struct {
	DryRun         bool   `json:"dry_run"`
	WindowHours    int    `json:"window_hours"`
	Checked        int    `json:"checked"`
	Rejected       int    `json:"rejected"`
	NotifyFailures int    `json:"notify_failures"`
	Message        string `json:"message,omitempty"`
}

// AutoRejectStaleEnquiries is the HTTP entry-point invoked by Cloud Scheduler.
func AutoRejectStaleEnquiries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dryRun := strings.EqualFold(r.URL.Query().Get("dryRun"), "true")

	hours := autoRejectHours()
	if hours <= 0 {
		writeJSON(w, http.StatusOK, runSummary{
			DryRun: dryRun, WindowHours: hours,
			Message: "ENQUIRY_AUTO_REJECT_HOURS is 0/unset — sweep disabled",
		})
		return
	}
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)

	projectID := resolveProjectID()
	notification.InitSharedFirebaseApp(projectID, "")
	fs, err := notification.SharedFirestoreClient(ctx)
	if err != nil {
		log.Printf("ERROR: firestore client: %v", err)
		writeJSON(w, http.StatusInternalServerError, runSummary{Message: fmt.Sprintf("firestore client: %v", err)})
		return
	}

	svc, err := notification.NewService(ctx, notification.Config{
		ProjectID:                     projectID,
		FCMTokenCollection:            "fcmTokens",
		NotificationHistoryCollection: "notificationHistory",
	})
	if err != nil {
		log.Printf("ERROR: notification service: %v", err)
		writeJSON(w, http.StatusInternalServerError, runSummary{Message: fmt.Sprintf("notification service: %v", err)})
		return
	}
	defer svc.Close()

	sum := runSummary{DryRun: dryRun, WindowHours: hours}
	for _, b := range buckets {
		sweepBucket(ctx, fs, svc, b, cutoff, dryRun, &sum)
	}

	log.Printf("enquiry-autoreject done: dryRun=%t checked=%d rejected=%d notifyFailures=%d",
		sum.DryRun, sum.Checked, sum.Rejected, sum.NotifyFailures)
	writeJSON(w, http.StatusOK, sum)
}

func sweepBucket(
	ctx context.Context,
	fs *firestore.Client,
	svc *notification.Service,
	b statusBucket,
	cutoff time.Time,
	dryRun bool,
	sum *runSummary,
) {
	iter := fs.Collection(enquiryCollection).Where("status", "in", b.statuses).Documents(ctx)
	defer iter.Stop()
	for {
		docSnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("WARN: %s sweep: iterate: %v", b.label, err)
			break
		}
		sum.Checked++

		if docSnap.UpdateTime.After(cutoff) {
			continue // still within the grace period
		}

		data := docSnap.Data()
		userID := docString(data, "userId", "user_id")
		sellerID := docString(data, "sellerId", "seller_id", "shopId", "shop_id")

		log.Printf("%s: %s (last updated %s) → reject: %s",
			b.label, docSnap.Ref.ID, docSnap.UpdateTime.Format(time.RFC3339), b.reason)

		if dryRun {
			sum.Rejected++ // "would reject" count
			continue
		}

		_, err = docSnap.Ref.Update(ctx, []firestore.Update{
			{Path: "status", Value: autoRejectStatus},
			{Path: "finalStatus", Value: autoRejectStatus},
			{Path: "rejectedBy", Value: b.rejectedBy},
			{Path: "autoRejectReason", Value: b.reason},
			{Path: "autoRejectedAt", Value: firestore.ServerTimestamp},
		})
		if err != nil {
			log.Printf("WARN: reject failed for %s: %v", docSnap.Ref.ID, err)
			continue
		}
		sum.Rejected++

		// Notify both parties — neither acted, so both should know it closed.
		notifyOwner(ctx, svc, "users", userID, b.reason, docSnap.Ref.ID, sum)
		notifyOwner(ctx, svc, "sellers", sellerID, b.reason, docSnap.Ref.ID, sum)
	}
}

func notifyOwner(
	ctx context.Context,
	svc *notification.Service,
	collection, ownerID, reason, enquiryID string,
	sum *runSummary,
) {
	if ownerID == "" {
		return
	}
	tokens, err := svc.GetOwnerFCMTokens(ctx, collection, ownerID)
	if err != nil || len(tokens) == 0 {
		if err != nil {
			log.Printf("WARN: token fetch %s/%s: %v", collection, ownerID, err)
		}
		return
	}
	data := map[string]string{"event_type": "enquiry_auto_rejected", "enquiry_id": enquiryID}
	if err := svc.SendToTokens(ctx, tokens, "Enquiry auto-rejected", reason, data); err != nil {
		sum.NotifyFailures++
		log.Printf("WARN: notify %s/%s failed: %v", collection, ownerID, err)
	}
}

func autoRejectHours() int {
	n, _ := strconv.Atoi(os.Getenv("ENQUIRY_AUTO_REJECT_HOURS"))
	return n
}

func resolveProjectID() string {
	if v := os.Getenv("GCP_PROJECT"); v != "" {
		return v
	}
	return os.Getenv("GOOGLE_CLOUD_PROJECT")
}

// docString returns the first present, non-empty string among keys — enquiry
// documents have used both camelCase and snake_case field names over time.
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

func writeJSON(w http.ResponseWriter, status int, body runSummary) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
