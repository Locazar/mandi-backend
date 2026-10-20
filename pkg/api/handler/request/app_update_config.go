package request

// UpdateAppUpdateConfig edits the seller/customer app force-update config.
type UpdateAppUpdateConfig struct {
	MinimumAppVersion string `json:"minimum_app_version"`
	LatestVersion     string `json:"latest_version"`
	ForceUpdate       bool   `json:"force_update"`
	PlayStoreURL      string `json:"play_store_url"`
	AppStoreURL       string `json:"app_store_url"`
}
