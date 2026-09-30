package domain

import "time"

// Verification notification template keys — one per distinct seller-facing
// push a document-verification action can trigger. Fixed set, never
// created/deleted through the API — only their Title/Body/ImageURL/Route
// change (same contract as OnboardingNudgeTemplate).
const (
	// VerifKeyFullyVerified fires from VerifyShop when all four checks
	// (shop photo, shop address, business doc, identity doc) are verified.
	VerifKeyFullyVerified = "shop_fully_verified"
	// VerifKeyLivePartial fires from VerifyShop when both mandatory checks
	// (shop photo + shop address) pass but at least one optional document
	// (business/identity) is still pending.
	VerifKeyLivePartial = "shop_live_partial"
	// VerifKeyPendingReview fires from VerifyShop when a mandatory check
	// hasn't passed yet — including when an admin un-verifies a check that
	// had previously passed, dropping the shop back to under_review.
	VerifKeyPendingReview = "shop_pending_review"
	// VerifKeyApproved fires from the separate ApproveShop action (the
	// admin's explicit go-live decision, distinct from a per-document
	// VerifyShop save).
	VerifKeyApproved = "shop_approved"
	// VerifKeyRejected fires from RejectShop.
	VerifKeyRejected = "shop_rejected"
)

// VerificationNotificationTemplateOrder lists every key, for listing/seeding.
var VerificationNotificationTemplateOrder = []string{
	VerifKeyFullyVerified, VerifKeyLivePartial, VerifKeyPendingReview,
	VerifKeyApproved, VerifKeyRejected,
}

// VerificationNotificationTemplate is editable copy for one shop
// document-verification push. Placeholders substituted at send time:
//
//	{{shop_name}}             — the shop's name
//	{{photo_status}}          — "verified ✓" or "pending ✗" (VerifyShop keys only)
//	{{address_status}}        — same, for shop address
//	{{business_doc_status}}   — same, for business document
//	{{identity_doc_status}}   — same, for identity document
//	{{remark}}                — the admin's rejection remark (VerifKeyRejected only;
//	                            substituted with "" when no remark was given)
//
// Route mirrors OnboardingNudgeTemplate.Route: where tapping the notification
// lands in the seller app. Only argument-free destinations make sense here —
// one template's route applies to every shop it's ever sent to.
type VerificationNotificationTemplate struct {
	Key       string    `json:"key" gorm:"primaryKey;type:varchar(32)"`
	Title     string    `json:"title" gorm:"type:varchar(200);not null"`
	Body      string    `json:"body" gorm:"type:text;not null"`
	ImageURL  string    `json:"image_url" gorm:"type:text"`
	Route     string    `json:"route" gorm:"type:varchar(200)"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
