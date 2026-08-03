package domain

// UserType identifies the principal kind across auth/session structs.
type UserType string

const (
	UserTypeUser  UserType = "user"
	UserTypeAdmin UserType = "admin"
)

func (t UserType) IsValid() bool {
	switch t {
	case UserTypeUser, UserTypeAdmin:
		return true
	}
	return false
}

// ShopType / ShopStatus / ShopDocumentType — Shop classification.
type ShopType string

const (
	ShopTypeRetail    ShopType = "retail"
	ShopTypeWholesale ShopType = "wholesale"
	ShopTypeService   ShopType = "service"
)

func (s ShopType) IsValid() bool {
	switch s {
	case ShopTypeRetail, ShopTypeWholesale, ShopTypeService:
		return true
	}
	return false
}

type ShopStatusType string

const (
	ShopStatusActive      ShopStatusType = "active"
	ShopStatusInactive    ShopStatusType = "inactive"
	ShopStatusSuspended   ShopStatusType = "suspended"
	ShopStatusUnderReview ShopStatusType = "under_review"
	ShopStatusRejected    ShopStatusType = "rejected"
)

func (s ShopStatusType) IsValid() bool {
	switch s {
	case ShopStatusActive, ShopStatusInactive, ShopStatusSuspended, ShopStatusUnderReview, ShopStatusRejected:
		return true
	}
	return false
}

type ShopDocumentType string

const (
	ShopDocGST     ShopDocumentType = "gst"
	ShopDocPAN     ShopDocumentType = "pan"
	ShopDocAadhaar ShopDocumentType = "aadhaar"
	ShopDocLicense ShopDocumentType = "license"
)

func (d ShopDocumentType) IsValid() bool {
	switch d {
	case ShopDocGST, ShopDocPAN, ShopDocAadhaar, ShopDocLicense:
		return true
	}
	return false
}

// OfferType — Offer.OfferType.
type OfferType string

const (
	OfferTypePercentage OfferType = "percentage"
	OfferTypeFixed      OfferType = "fixed"
)

func (o OfferType) IsValid() bool {
	switch o {
	case OfferTypePercentage, OfferTypeFixed:
		return true
	}
	return false
}

// FieldType — SubTypeAttributes.FieldType.
type FieldType string

const (
	FieldTypeDropdown FieldType = "dropdown"
	FieldTypeNumber   FieldType = "number"
	FieldTypeText     FieldType = "text"
)

func (f FieldType) IsValid() bool {
	switch f {
	case FieldTypeDropdown, FieldTypeNumber, FieldTypeText:
		return true
	}
	return false
}

// AddressType — Address.AddressType.
type AddressType string

const (
	AddressTypeHome  AddressType = "home"
	AddressTypeWork  AddressType = "work"
	AddressTypeOther AddressType = "other"
)

func (a AddressType) IsValid() bool {
	switch a {
	case AddressTypeHome, AddressTypeWork, AddressTypeOther:
		return true
	}
	return false
}

// OfferEventType / ExperimentVariant — UserOfferHistory.
type OfferEventType string

const (
	OfferEventShown     OfferEventType = "shown"
	OfferEventClicked   OfferEventType = "clicked"
	OfferEventDismissed OfferEventType = "dismissed"
	OfferEventApplied   OfferEventType = "applied"
)

func (e OfferEventType) IsValid() bool {
	switch e {
	case OfferEventShown, OfferEventClicked, OfferEventDismissed, OfferEventApplied:
		return true
	}
	return false
}

// AlertType / AlertActionType / AlertTemplateType / AlertDisplayType.
type AlertType string

const (
	AlertTypeInfo     AlertType = "info"
	AlertTypeWarning  AlertType = "warning"
	AlertTypeCritical AlertType = "critical"
)

func (a AlertType) IsValid() bool {
	switch a {
	case AlertTypeInfo, AlertTypeWarning, AlertTypeCritical:
		return true
	}
	return false
}

type AlertActionType string

const (
	AlertActionNavigate AlertActionType = "navigate"
	AlertActionAPICall  AlertActionType = "api_call"
	AlertActionDismiss  AlertActionType = "dismiss"
)

func (a AlertActionType) IsValid() bool {
	switch a {
	case AlertActionNavigate, AlertActionAPICall, AlertActionDismiss:
		return true
	}
	return false
}

type AlertDisplayType string

const (
	AlertDisplayBottomSheet AlertDisplayType = "bottom_sheet"
	AlertDisplayBanner      AlertDisplayType = "banner"
	AlertDisplayModal       AlertDisplayType = "modal"
)

func (a AlertDisplayType) IsValid() bool {
	switch a {
	case AlertDisplayBottomSheet, AlertDisplayBanner, AlertDisplayModal:
		return true
	}
	return false
}

// NotificationStatus — Notification.Status. (SenderType/ReceiverType reuse UserType
// semantics but keep a dedicated type for the sender/receiver columns.)
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusRead    NotificationStatus = "read"
	NotificationStatusFailed  NotificationStatus = "failed"
)

func (s NotificationStatus) IsValid() bool {
	switch s {
	case NotificationStatusPending, NotificationStatusSent, NotificationStatusRead, NotificationStatusFailed:
		return true
	}
	return false
}

// MobileAuthStatus — OTPRequest.Status (mobile_auth.go).
type MobileAuthStatus string

const (
	MobileAuthActive   MobileAuthStatus = "active"
	MobileAuthVerified MobileAuthStatus = "verified"
	MobileAuthExpired  MobileAuthStatus = "expired"
	MobileAuthBlocked  MobileAuthStatus = "blocked"
)

func (s MobileAuthStatus) IsValid() bool {
	switch s {
	case MobileAuthActive, MobileAuthVerified, MobileAuthExpired, MobileAuthBlocked:
		return true
	}
	return false
}

// ShopTimeStatus — ShopTime.Status.
type ShopTimeStatus string

const (
	ShopTimeOpen  ShopTimeStatus = "open"
	ShopTimeClose ShopTimeStatus = "close"
)

func (s ShopTimeStatus) IsValid() bool {
	switch s {
	case ShopTimeOpen, ShopTimeClose:
		return true
	}
	return false
}

// BannerAppType — Banner.AppType.
type BannerAppType string

const (
	BannerAppCustomer BannerAppType = "customer"
	BannerAppSeller   BannerAppType = "seller"
)

func (b BannerAppType) IsValid() bool {
	switch b {
	case BannerAppCustomer, BannerAppSeller:
		return true
	}
	return false
}

// SubscriptionStatus — SubscriptionOrder.Status (replaces the loose string consts in subscription.go).
type SubscriptionStatus string

const (
	SubscriptionStatusCreated SubscriptionStatus = "created"
	SubscriptionStatusPaid    SubscriptionStatus = "paid"
	SubscriptionStatusFailed  SubscriptionStatus = "failed"
	SubscriptionStatusExpired SubscriptionStatus = "expired"
)

func (s SubscriptionStatus) IsValid() bool {
	switch s {
	case SubscriptionStatusCreated, SubscriptionStatusPaid, SubscriptionStatusFailed, SubscriptionStatusExpired:
		return true
	}
	return false
}

// ConsentType — UserConsent.ConsentType (Phase 6).
type ConsentType string

const (
	ConsentTypeTerms     ConsentType = "terms"
	ConsentTypePrivacy   ConsentType = "privacy"
	ConsentTypeMarketing ConsentType = "marketing"
)

func (c ConsentType) IsValid() bool {
	switch c {
	case ConsentTypeTerms, ConsentTypePrivacy, ConsentTypeMarketing:
		return true
	}
	return false
}

// FirestoreEventType — FirestoreEvent recipient type.
type FirestoreEventType string

const (
	FirestoreEventUser  FirestoreEventType = "user"
	FirestoreEventAdmin FirestoreEventType = "admin"
)

func (t FirestoreEventType) IsValid() bool {
	switch t {
	case FirestoreEventUser, FirestoreEventAdmin:
		return true
	}
	return false
}
