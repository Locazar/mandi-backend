// Command fcmbackfill copies active FCM device tokens from Postgres
// (notification_device_tokens) into Firestore under the owner documents that
// the enquiry Cloud Function actually looks up:
//
//	seller tokens → sellers/{shop_id}/fcmTokens/{token}
//	user tokens   → users/{owner_id}/fcmTokens/{token}
//
// Why this exists: historically seller tokens were synced to Firestore under
// the admin id (adm_...) or legacy numeric ids, while enquiry documents
// reference sellers by shop id (shp_...). The Cloud Function therefore found
// no tokens and no push was delivered. Run this once (idempotent — safe to
// re-run) after deploying the fixed backend so existing devices get healed
// without requiring an app update or re-login.
//
// Usage (reads DB_* and FIREBASE_* settings from the same env/.env as the API):
//
//	go run ./cmd/fcmbackfill            # dry-run: prints what would be written
//	go run ./cmd/fcmbackfill -apply     # actually writes to Firestore
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/rohit221990/mandi-backend/pkg/config"
	notificationSvc "github.com/rohit221990/mandi-backend/pkg/service/notification"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type deviceToken struct {
	OwnerID   string
	ShopID    string
	OwnerType string
	Token     string
	Platform  string
}

func main() {
	apply := flag.Bool("apply", false, "write to Firestore (default is dry-run)")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("load config: ", err)
	}

	// Plain connection — deliberately NOT db.ConnectDatabase, which also runs
	// schema migrations; a read-only backfill must not mutate the schema.
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%s password=%s", cfg.DBHost, cfg.DBUser, cfg.DBName, cfg.DBPort, cfg.DBPassword)
	if cfg.DBSSLRootCert != "" {
		dsn += fmt.Sprintf(" sslmode=verify-full sslrootcert=%s", cfg.DBSSLRootCert)
	} else {
		dsn += " sslmode=disable"
	}
	dbConn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("connect db: ", err)
	}

	var rows []deviceToken
	if err := dbConn.Raw(`SELECT owner_id, shop_id, owner_type, token, platform
		FROM notification_device_tokens WHERE is_active = TRUE`).Scan(&rows).Error; err != nil {
		log.Fatal("query notification_device_tokens: ", err)
	}
	log.Printf("found %d active device token(s) in Postgres", len(rows))

	notificationSvc.InitSharedFirebaseApp(cfg.FirebaseProjectID, cfg.FirebaseConfig)
	fs, err := notificationSvc.SharedFirestoreClient(context.Background())
	if err != nil {
		log.Fatal("firestore client: ", err)
	}
	defer fs.Close()

	ctx := context.Background()
	written, skipped := 0, 0
	for _, r := range rows {
		collection := "users"
		docID := r.OwnerID
		if r.OwnerType == "seller" {
			collection = "sellers"
			if r.ShopID != "" {
				docID = r.ShopID
			}
		}
		if docID == "" || r.Token == "" {
			skipped++
			continue
		}

		short := r.Token
		if len(short) > 16 {
			short = short[:16] + "..."
		}
		fmt.Printf("%s/%s ← token=%s platform=%s\n", collection, docID, short, r.Platform)

		if !*apply {
			continue
		}
		_, err := fs.Collection(collection).Doc(docID).Collection("fcmTokens").Doc(r.Token).Set(ctx, map[string]interface{}{
			"token":     r.Token,
			"platform":  r.Platform,
			"isActive":  true,
			"updatedAt": firestore.ServerTimestamp,
		}, firestore.MergeAll)
		if err != nil {
			log.Printf("WARN: write failed for %s/%s: %v", collection, docID, err)
			continue
		}
		written++
	}

	if *apply {
		log.Printf("done: %d written, %d skipped", written, skipped)
	} else {
		log.Printf("dry-run complete (%d candidate(s), %d skipped) — re-run with -apply to write", len(rows)-skipped, skipped)
	}
}
