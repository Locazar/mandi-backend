package notification

// firebase_app.go provides a package-level singleton Firebase App so that
// FCMPushService and FirestoreWatcher share exactly ONE firebase.App instance.
// The Firebase Admin Go SDK returns an error if you call firebase.NewApp with
// the same (default) name more than once, which silently breaks the second
// caller.  Using this singleton avoids that.

import (
	"context"
	"fmt"
	"os"
	"sync"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var (
	sharedApp     *firebase.App
	sharedAppOnce sync.Once
	sharedAppErr  error
)

// getSharedApp returns the single shared Firebase App, initialising it on the
// first call.  Subsequent calls return the cached instance (or cached error).
func getSharedApp(ctx context.Context) (*firebase.App, error) {
	sharedAppOnce.Do(func() {
		// Project ID: read from env, fall back to the known project ID.
		projectID := os.Getenv("FIREBASE_PROJECT_ID")
		if projectID == "" {
			projectID = "locazar-f20b6"
		}
		conf := &firebase.Config{ProjectID: projectID}

		credJSON := os.Getenv("FIREBASE_CONFIG")
		if credJSON != "" {
			opt := option.WithCredentialsJSON([]byte(credJSON))
			sharedApp, sharedAppErr = firebase.NewApp(ctx, conf, opt)
		} else {
			// Falls back to GOOGLE_APPLICATION_CREDENTIALS file or Workload Identity
			sharedApp, sharedAppErr = firebase.NewApp(ctx, conf)
		}
		if sharedAppErr != nil {
			sharedAppErr = fmt.Errorf("firebase app init: %w", sharedAppErr)
		}
	})
	return sharedApp, sharedAppErr
}

// sharedMessagingClient returns the FCM Messaging client from the shared app.
func sharedMessagingClient(ctx context.Context) (*messaging.Client, error) {
	app, err := getSharedApp(ctx)
	if err != nil {
		return nil, err
	}
	return app.Messaging(ctx)
}

// sharedFirestoreClient returns the Firestore client from the shared app.
func sharedFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	app, err := getSharedApp(ctx)
	if err != nil {
		return nil, err
	}
	return app.Firestore(ctx)
}
