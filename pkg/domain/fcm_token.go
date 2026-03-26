package domain

type FcmToken struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Token    string `gorm:"unique;not null" json:"token"`
	Device   string `json:"device"`
	Platform string `json:"platform"`
	ShopID   uint   `json:"shop_id"`
	AdminID  uint   `json:"admin_id"`
}
