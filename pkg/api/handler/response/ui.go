package response

// UITemplate represents the UI template response for different intents
type UITemplate struct {
	Intent   string      `json:"intent"`
	Title    string      `json:"title"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data"`
	Status   string      `json:"status"`
	Required []string    `json:"required,omitempty"`
}

// ShopImageTemplate for shop_image intent
type ShopImageTemplate struct {
	ShopID       uint        `json:"shop_id"`
	ShopName     string      `json:"shop_name"`
	CurrentImage string      `json:"current_image"`
	Status       string      `json:"status"`
	Message      string      `json:"message"`
	Action       string      `json:"action"`
	UploadURL    string      `json:"upload_url"`
}

// ShopVerificationTemplate for shop_verification intent
type ShopVerificationTemplate struct {
	ShopID              uint   `json:"shop_id"`
	ShopName            string `json:"shop_name"`
	VerificationStatus  string `json:"verification_status"`
	DocumentStatus      string `json:"document_status"`
	AddressVerified     bool   `json:"address_verified"`
	ImageVerified       bool   `json:"image_verified"`
	PendingItems        []string `json:"pending_items"`
	NextStep            string `json:"next_step"`
}

// AddProductTemplate for add_product intent
type AddProductTemplate struct {
	ShopID       uint          `json:"shop_id"`
	ShopName     string        `json:"shop_name"`
	Category     string        `json:"category"`
	RequiredFields []string    `json:"required_fields"`
	Message      string        `json:"message"`
	Action       string        `json:"action"`
	UploadURL    string        `json:"upload_url"`
}

// ProductImageTemplate for product_image intent
type ProductImageTemplate struct {
	ShopID        uint   `json:"shop_id"`
	ProductItemID string `json:"product_item_id"`
	ProductName   string `json:"product_name"`
	CurrentImage  string `json:"current_image"`
	Status        string `json:"status"`
	Message       string `json:"message"`
	Action        string `json:"action"`
	UploadURL     string `json:"upload_url"`
}

// ShopAddressTemplate for shop_address intent
type ShopAddressTemplate struct {
	ShopID    uint   `json:"shop_id"`
	ShopName  string `json:"shop_name"`
	Address   string `json:"address"`
	City      string `json:"city"`
	State     string `json:"state"`
	Pincode   string `json:"pincode"`
	Verified  bool   `json:"verified"`
	Message   string `json:"message"`
	Action    string `json:"action"`
	UploadURL string `json:"upload_url"`
}
