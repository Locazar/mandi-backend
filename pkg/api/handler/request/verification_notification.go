package request

// UpdateVerificationNotificationTemplate edits one of the five fixed
// document-verification template keys. Key selects which one (from the URL
// path); it is never itself editable.
type UpdateVerificationNotificationTemplate struct {
	Title    string `json:"title" binding:"required"`
	Body     string `json:"body" binding:"required"`
	ImageURL string `json:"image_url"`
	// Route is optional — an empty value means the tap just opens the app.
	Route string `json:"route"`
}
