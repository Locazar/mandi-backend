package domain

// AppUpdatePlatform selects which app's update config a request targets.
type AppUpdatePlatform string

const (
	AppUpdatePlatformSeller   AppUpdatePlatform = "seller"
	AppUpdatePlatformCustomer AppUpdatePlatform = "customer"
)

// IsValid reports whether p is one of the two known platforms — used to
// reject anything else before it ever reaches a Firestore path, since the
// platform becomes part of the document path (config_{platform}/app).
func (p AppUpdatePlatform) IsValid() bool {
	return p == AppUpdatePlatformSeller || p == AppUpdatePlatformCustomer
}

// FirestoreCollection returns the config_seller / config_customer collection
// this platform's update doc lives in.
func (p AppUpdatePlatform) FirestoreCollection() string {
	return "config_" + string(p)
}

// AppUpdateConfig is the config_{platform}/app Firestore document shape,
// read directly by the seller/customer app at startup to decide whether to
// force an update (see seller-app's AppUpdateService.check()). Field names
// match the Firestore document keys exactly (camelCase) — these are read
// as-is by Dart/Kotlin client code, not converted through Go's usual
// snake_case JSON convention.
//
// Only these 5 fields are ever written by SaveAppUpdateConfig, via a
// Firestore merge (not a full document replace) — config_seller/app also
// carries privacyPolicyUrl/termsOfServiceUrl (read by the same client
// service) that this type intentionally doesn't model, so saving here can
// never accidentally wipe them.
type AppUpdateConfig struct {
	MinimumAppVersion string `json:"minimum_app_version" firestore:"minimumAppVersion"`
	LatestVersion     string `json:"latest_version" firestore:"latestVersion"`
	ForceUpdate       bool   `json:"force_update" firestore:"forceUpdate"`
	PlayStoreURL      string `json:"play_store_url" firestore:"playStoreUrl"`
	AppStoreURL       string `json:"app_store_url" firestore:"appStoreUrl"`
}
