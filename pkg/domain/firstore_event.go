package domain

// FirestoreEvent represents the Eventarc Firestore event structure
// Reference: google.cloud.firestore.document.v1.updated
type FirestoreEvent struct {
	Data FirestoreEventData `json:"data"`
	ID   string             `json:"id"`
}

// FirestoreEventData contains the actual Firestore document update data
type FirestoreEventData struct {
	Value      *FirestoreDocument `json:"value"`
	OldValue   *FirestoreDocument `json:"oldValue"`
	UpdateMask *UpdateMask        `json:"updateMask"`
}

// FirestoreDocument represents a Firestore document with typed fields
type FirestoreDocument struct {
	Name       string                 `json:"name"`       // Full resource name: projects/{project}/databases/(default)/documents/{path}
	Fields     map[string]interface{} `json:"fields"`     // Typed field values
	CreateTime string                 `json:"createTime"` // RFC 3339 format
	UpdateTime string                 `json:"updateTime"` // RFC 3339 format
	DeleteTime string                 `json:"deleteTime"` // RFC 3339 format (if deleted)
}

// UpdateMask indicates which fields were updated
type UpdateMask struct {
	FieldPaths []string `json:"fieldPaths"`
}

// ParsedFirestoreEvent represents a structured representation of the event
type ParsedFirestoreEvent struct {
	DocumentPath string                 // e.g., "enquiries/doc-id"
	DocumentID   string                 // e.g., "doc-id"
	UpdateTime   string                 // RFC 3339 formatted timestamp
	OldFields    map[string]interface{} // Parsed old field values
	NewFields    map[string]interface{} // Parsed new field values
	UpdatedPaths []string               // Fields that were updated
}

// FieldChange represents a change to a monitored field
type FieldChange struct {
	FieldName string      // Name of the field
	OldValue  interface{} // Previous value (nil if new)
	NewValue  interface{} // New value (nil if deleted)
}

// NotificationPayload contains data for the FCM notification
type NotificationPayload struct {
	DocumentID    string   // Firestore document ID
	DocumentPath  string   // Full document path
	EnquiryID     string   // Enquiry/Query ID
	UserID        string   // User (buyer) who created the enquiry
	SellerID      string   // Seller/shop that owns the enquiry
	AssignedTo    string   // Support agent assigned to handle enquiry
	Title         string   // Notification title
	Body          string   // Notification body
	ChangeCount   int      // Number of fields changed
	ChangedFields []string // List of changed field names
	Timestamp     string   // Update timestamp
	ActionURL     string   // Deep-link URL for the app
}

// NotificationRecipient represents a user who should receive the notification
type NotificationRecipient struct {
	UserID string   // User identifier
	Tokens []string // FCM tokens
	Type   string   // "user" or "admin"
}

// EnquiryStatus represents different enquiry states
type EnquiryStatus string

const (
	// Active / in-flight states
	StatusNew                 EnquiryStatus = "new"
	StatusInProgress          EnquiryStatus = "in_progress"
	StatusPendingSellerPrice  EnquiryStatus = "pending_seller_price"
	StatusPendingCustomPrice  EnquiryStatus = "pending_customer_price"
	StatusPendingSellerFinal  EnquiryStatus = "pending_seller_final"
	StatusPendingCustomFinal  EnquiryStatus = "pending_customer_final"
	StatusSellerFinalUpdate   EnquiryStatus = "seller_final_update"
	StatusCounterOffer        EnquiryStatus = "counter_offer"
	StatusOnHold              EnquiryStatus = "on_hold"
	StatusCustomerFinalUpdate EnquiryStatus = "customer_accepted_final"

	// Terminal / resolved states
	StatusCompletedAccepted EnquiryStatus = "completed_accepted"
	StatusCompletedRejected EnquiryStatus = "completed_rejected"
	StatusResolved          EnquiryStatus = "resolved"
	StatusClosed            EnquiryStatus = "closed"
	StatusRejected          EnquiryStatus = "rejected"
	StatusCancelled         EnquiryStatus = "cancelled"
	StatusExpired           EnquiryStatus = "expired"
	StatusReopened          EnquiryStatus = "reopened"

	// Dispute states
	StatusDispute         EnquiryStatus = "dispute"
	StatusDisputeResolved EnquiryStatus = "dispute_resolved"
)
