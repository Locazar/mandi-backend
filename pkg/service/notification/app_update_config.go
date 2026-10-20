package notification

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// GetAppUpdateConfig reads config_{platform}/app — the same document the
// seller/customer app itself reads at startup (see AppUpdateService.check()
// in seller-app). A missing document (first run, nothing saved yet) returns
// a zero-value config and no error, matching that client's own "fails open"
// contract — there's simply nothing to show yet.
func GetAppUpdateConfig(ctx context.Context, platform domain.AppUpdatePlatform) (domain.AppUpdateConfig, error) {
	client, err := SharedFirestoreClient(ctx)
	if err != nil {
		return domain.AppUpdateConfig{}, fmt.Errorf("firestore client: %w", err)
	}
	snap, err := client.Collection(platform.FirestoreCollection()).Doc("app").Get(ctx)
	if err != nil {
		if grpcstatus.Code(err) == codes.NotFound {
			return domain.AppUpdateConfig{}, nil
		}
		return domain.AppUpdateConfig{}, err
	}
	var cfg domain.AppUpdateConfig
	if err := snap.DataTo(&cfg); err != nil {
		return domain.AppUpdateConfig{}, fmt.Errorf("decode app update config: %w", err)
	}
	return cfg, nil
}

// SaveAppUpdateConfig merges cfg's 5 fields into config_{platform}/app —
// deliberately a merge, not a full-document Set, so config_seller/app's
// other fields (privacyPolicyUrl, termsOfServiceUrl — read by the same
// client service, never modeled here) are never touched by this write.
func SaveAppUpdateConfig(ctx context.Context, platform domain.AppUpdatePlatform, cfg domain.AppUpdateConfig) error {
	client, err := SharedFirestoreClient(ctx)
	if err != nil {
		return fmt.Errorf("firestore client: %w", err)
	}
	_, err = client.Collection(platform.FirestoreCollection()).Doc("app").Set(ctx, map[string]interface{}{
		"minimumAppVersion": cfg.MinimumAppVersion,
		"latestVersion":     cfg.LatestVersion,
		"forceUpdate":       cfg.ForceUpdate,
		"playStoreUrl":      cfg.PlayStoreURL,
		"appStoreUrl":       cfg.AppStoreURL,
	}, firestore.MergeAll)
	return err
}
