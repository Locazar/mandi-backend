package domain

// ShopAnnouncementTarget is the slice of a shop row a launch announcement needs:
// who it is, where it is, and what picture to show.
type ShopAnnouncementTarget struct {
	ID           string  `json:"id" gorm:"column:id"`
	ShopName     string  `json:"shop_name" gorm:"column:shop_name"`
	City         string  `json:"city" gorm:"column:city"`
	Latitude     float64 `json:"latitude" gorm:"column:latitude"`
	Longitude    float64 `json:"longitude" gorm:"column:longitude"`
	ShopImageURL string  `json:"shop_image_url" gorm:"column:shop_image_url"`
	ShopStatus   string  `json:"shop_status" gorm:"column:shop_status"`
}

// HasLocation reports whether the shop has a usable pinned position. A shop
// saved without one lands at (0, 0) — a point in the Atlantic — which would
// match no customer at all, so callers must reject it rather than send a push
// that silently reaches nobody.
func (s ShopAnnouncementTarget) HasLocation() bool {
	return s.Latitude != 0 && s.Longitude != 0
}

// CustomerDevice is one reachable device: which customer owns it and the FCM
// token to deliver to. One customer may contribute several rows (phone, tablet).
type CustomerDevice struct {
	UserID string `json:"user_id" gorm:"column:user_id"`
	Token  string `json:"token" gorm:"column:token"`
}

// RadiusAnnouncementResult reports what a send actually did, so the admin portal
// can show real numbers instead of a bare "sent" toast. ShopID/ShopName are set
// only for shop-launch announcements; a raw-coordinate geo push leaves them empty.
type RadiusAnnouncementResult struct {
	ShopID        string  `json:"shop_id,omitempty"`
	ShopName      string  `json:"shop_name,omitempty"`
	RadiusKm      float64 `json:"radius_km"`
	CustomerCount int     `json:"customer_count"`
	DeviceCount   int     `json:"device_count"`
	// DeliveredDevices counts tokens FCM accepted. Lower than DeviceCount when
	// some devices have been uninstalled since their token was registered.
	DeliveredDevices int    `json:"delivered_devices"`
	ImageURL         string `json:"image_url"`
}
