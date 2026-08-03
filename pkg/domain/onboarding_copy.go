package domain

// OnboardingWizardCopyKey is the fixed app_configs.config_key under which the
// single structured OnboardingWizardCopy JSON blob is stored (see
// GET/PUT /api/admin/onboarding-copy and the public GET /api/onboarding-copy).
const OnboardingWizardCopyKey = "onboarding_wizard_copy"

// OnboardingWizardCopy holds every admin-editable text string shown across
// the seller-app's shop-onboarding wizard (title/subtitle/button labels/field
// labels/hints/messages). It does NOT cover field structure, order,
// validation rules, or business logic (e.g. the phone-visibility consent
// behavior itself) — those stay fixed in the seller-app's code. Stored as a
// single JSON blob in the app_configs table under OnboardingWizardCopyKey.
type OnboardingWizardCopy struct {
	ShopIdentity OnboardingShopIdentityCopy `json:"shopIdentity"`
	Location     OnboardingLocationCopy     `json:"location"`
	Category     OnboardingCategoryCopy     `json:"category"`
	Documents    OnboardingDocumentsCopy    `json:"documents"`
	ShopPhoto    OnboardingShopPhotoCopy    `json:"shopPhoto"`
	Review       OnboardingReviewCopy       `json:"review"`
}

type OnboardingShopIdentityCopy struct {
	Title                  string `json:"title"`
	Subtitle               string `json:"subtitle"`
	PrimaryLabel           string `json:"primaryLabel"`
	ShopNameLabel          string `json:"shopNameLabel"`
	ShopNameHint           string `json:"shopNameHint"`
	ShopNameTooShortError  string `json:"shopNameTooShortError"`
	OwnerNameLabel         string `json:"ownerNameLabel"`
	OwnerNameHint          string `json:"ownerNameHint"`
	OwnerNameTooShortError string `json:"ownerNameTooShortError"`
	PhoneFieldTitle        string `json:"phoneFieldTitle"`
	PhoneVisibleBadge      string `json:"phoneVisibleBadge"`
	PhoneConsentLabel      string `json:"phoneConsentLabel"`
	HideNumberDialogTitle  string `json:"hideNumberDialogTitle"`
	HideNumberDialogBody   string `json:"hideNumberDialogBody"`
	HideNumberKeepButton   string `json:"hideNumberKeepButton"`
	HideNumberHideButton   string `json:"hideNumberHideButton"`
}

type OnboardingLocationCopy struct {
	Title                    string `json:"title"`
	Subtitle                 string `json:"subtitle"`
	PrimaryLabel             string `json:"primaryLabel"`
	Address1Label            string `json:"address1Label"`
	Address1Hint             string `json:"address1Hint"`
	Address2Label            string `json:"address2Label"`
	Address2Hint             string `json:"address2Hint"`
	CityLabel                string `json:"cityLabel"`
	CityHint                 string `json:"cityHint"`
	PincodeLabel             string `json:"pincodeLabel"`
	PincodeHint              string `json:"pincodeHint"`
	PincodeDigitsOnlyError   string `json:"pincodeDigitsOnlyError"`
	PincodeLengthError       string `json:"pincodeLengthError"`
	StateDropdownHint        string `json:"stateDropdownHint"`
	LocationPinnedTitle      string `json:"locationPinnedTitle"`
	LocationUnpinnedTitle    string `json:"locationUnpinnedTitle"`
	LocationUnpinnedSubtitle string `json:"locationUnpinnedSubtitle"`
	PincodeLookupLoading     string `json:"pincodeLookupLoading"`
	PincodeMismatchPrefix    string `json:"pincodeMismatchPrefix"`
	PincodeMismatchSuffix    string `json:"pincodeMismatchSuffix"`
	GetLocationButton        string `json:"getLocationButton"`
	UpdateLocationButton     string `json:"updateLocationButton"`
	LocationSuccessToast     string `json:"locationSuccessToast"`
	// Shown as a confirm dialog before capturing GPS (onboarding + profile
	// update) so the seller only pins coordinates while physically at the shop.
	LocationConfirmTitle        string `json:"locationConfirmTitle"`
	LocationConfirmBody         string `json:"locationConfirmBody"`
	LocationConfirmConfirmLabel string `json:"locationConfirmConfirmLabel"`
	LocationConfirmCancelLabel  string `json:"locationConfirmCancelLabel"`
}

type OnboardingCategoryCopy struct {
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	PrimaryLabel string `json:"primaryLabel"`
}

type OnboardingDocumentsCopy struct {
	Title                   string `json:"title"`
	Subtitle                string `json:"subtitle"`
	SendOtpLabel            string `json:"sendOtpLabel"`
	ContinueLabel           string `json:"continueLabel"`
	SkipLabel               string `json:"skipLabel"`
	DocumentNumberLabel     string `json:"documentNumberLabel"`
	DocumentNumberHintPre   string `json:"documentNumberHintPrefix"`
	DocumentNumberHintPost  string `json:"documentNumberHintSuffix"`
	OtpSentToast            string `json:"otpSentToast"`
	DocumentVerifiedLabel   string `json:"documentVerifiedLabel"`
	EnterOtpLabel           string `json:"enterOtpLabel"`
	ResendOtpLabel          string `json:"resendOtpLabel"`
	ResendOtpCountdownTmpl  string `json:"resendOtpCountdownTemplate"`
}

type OnboardingShopPhotoCopy struct {
	Title                string `json:"title"`
	Subtitle             string `json:"subtitle"`
	PrimaryLabel         string `json:"primaryLabel"`
	CapturePlaceholder   string `json:"capturePlaceholder"`
	RetakeButton         string `json:"retakeButton"`
	TipsButton           string `json:"tipsButton"`
	CaptureErrorPrefix   string `json:"captureErrorPrefix"`
}

type OnboardingReviewCopy struct {
	Title                     string `json:"title"`
	Subtitle                  string `json:"subtitle"`
	PrimaryLabel              string `json:"primaryLabel"`
	ShopIdentityCardTitle     string `json:"shopIdentityCardTitle"`
	LocationCardTitle         string `json:"locationCardTitle"`
	StoreTypeCardTitle        string `json:"storeTypeCardTitle"`
	DocumentsCardTitle        string `json:"documentsCardTitle"`
	ShopPhotoCardTitle        string `json:"shopPhotoCardTitle"`
	DocumentsNotVerifiedLabel string `json:"documentsNotVerifiedLabel"`
	ShopPhotoAddedLabel       string `json:"shopPhotoAddedLabel"`
	ShopPhotoMissingLabel     string `json:"shopPhotoMissingLabel"`
	DirtyBadgeLabel           string `json:"dirtyBadgeLabel"`
	EditButton                string `json:"editButton"`
	ReferralTitle             string `json:"referralTitle"`
	ReferralLabel             string `json:"referralLabel"`
	ReferralHint              string `json:"referralHint"`
}

// DefaultOnboardingWizardCopy returns the current, previously-hardcoded
// seller-app strings — used whenever no admin override has been saved yet,
// so behavior is unchanged until someone actually edits it from admin-portal.
func DefaultOnboardingWizardCopy() OnboardingWizardCopy {
	return OnboardingWizardCopy{
		ShopIdentity: OnboardingShopIdentityCopy{
			Title:                  "What's your shop called?",
			Subtitle:               "This is how customers will see your shop on Locazar.",
			PrimaryLabel:           "Continue",
			ShopNameLabel:          "Shop name",
			ShopNameHint:           "Enter your shop/business name",
			ShopNameTooShortError:  "Shop name is too short",
			OwnerNameLabel:         "Owner name",
			OwnerNameHint:          "Enter the owner's full name",
			OwnerNameTooShortError: "Owner name is too short",
			PhoneFieldTitle:        "Verified phone number",
			PhoneVisibleBadge:      "Visible To Customer",
			PhoneConsentLabel:      "I agree that this number will be visible to customers, so they can discover and contact your shop.",
			HideNumberDialogTitle:  "Hide your number?",
			HideNumberDialogBody:   "Customers may find it difficult to reach you if you don't allow your number to be visible on your shop. Are you sure you want to hide it?",
			HideNumberKeepButton:   "Keep visible",
			HideNumberHideButton:   "Hide anyway",
		},
		Location: OnboardingLocationCopy{
			Title:                    "Where is your shop?",
			Subtitle:                 "Customers nearby will discover you from this address.",
			PrimaryLabel:             "Save & Continue",
			Address1Label:            "Address line 1",
			Address1Hint:             "House/Flat No, Building Name",
			Address2Label:            "Address line 2 (optional)",
			Address2Hint:             "Street, Locality, Area",
			CityLabel:                "City",
			CityHint:                 "City/Village/Town",
			PincodeLabel:             "PIN code",
			PincodeHint:              "6 digits",
			PincodeDigitsOnlyError:   "Only digits allowed",
			PincodeLengthError:       "PIN code must be 6 digits",
			StateDropdownHint:        "Select your state",
			LocationPinnedTitle:      "Location pinned",
			LocationUnpinnedTitle:    "Pin your shop location",
			LocationUnpinnedSubtitle: "We use your coordinates for accurate delivery and verification.",
			PincodeLookupLoading:     "Looking up your PIN code…",
			PincodeMismatchPrefix:    "This PIN code resolves to ",
			PincodeMismatchSuffix:    ". Please update your pinned location below to match.",
			GetLocationButton:        "Get current location",
			UpdateLocationButton:     "Update location",
			LocationSuccessToast:     "Location Update successfully!",
			LocationConfirmTitle:        "Are you at your shop right now?",
			LocationConfirmBody:         "This saves your current location as your shop's location, so nearby customers can find you. Please make sure you're standing at your shop before continuing.",
			LocationConfirmConfirmLabel: "Yes, I'm at my shop",
			LocationConfirmCancelLabel:  "Not now",
		},
		Category: OnboardingCategoryCopy{
			Title:        "What kind of store do you run?",
			Subtitle:     "Choose the category that best describes your shop.",
			PrimaryLabel: "Save & Continue",
		},
		Documents: OnboardingDocumentsCopy{
			Title:                  "Verify your business",
			Subtitle:               "Add a document to become a verified seller. You can also do this later.",
			SendOtpLabel:           "Send OTP",
			ContinueLabel:          "Continue",
			SkipLabel:              "Skip",
			DocumentNumberLabel:    "Document number",
			DocumentNumberHintPre:  "Enter your ",
			DocumentNumberHintPost: " number",
			OtpSentToast:           "OTP sent for verification.",
			DocumentVerifiedLabel:  "Document verified",
			EnterOtpLabel:          "Enter the 6-digit OTP",
			ResendOtpLabel:         "Resend OTP",
			ResendOtpCountdownTmpl: "Resend OTP in {seconds}s",
		},
		ShopPhoto: OnboardingShopPhotoCopy{
			Title:              "Add a photo of your shop",
			Subtitle:           "A clear storefront photo helps customers recognise you.",
			PrimaryLabel:       "Upload & Continue",
			CapturePlaceholder: "Tap to take a shop photo",
			RetakeButton:       "Retake photo",
			TipsButton:         "How to take a shop photo?",
			CaptureErrorPrefix: "Error capturing photo: ",
		},
		Review: OnboardingReviewCopy{
			Title:                     "Review your details",
			Subtitle:                  "Make sure everything looks right before you submit.",
			PrimaryLabel:              "Submit for verification",
			ShopIdentityCardTitle:     "Shop identity",
			LocationCardTitle:         "Location",
			StoreTypeCardTitle:        "Store type",
			DocumentsCardTitle:        "Documents",
			ShopPhotoCardTitle:        "Shop photo",
			DocumentsNotVerifiedLabel: "Not verified — an agent will call you",
			ShopPhotoAddedLabel:       "Photo added",
			ShopPhotoMissingLabel:     "No photo",
			DirtyBadgeLabel:           "updates on submit",
			EditButton:                "Edit",
			ReferralTitle:             "Referral ID (optional)",
			ReferralLabel:             "Referral ID",
			ReferralHint:              "Enter a referral code, if you have one",
		},
	}
}
