package domain

import (
	"gorm.io/gorm"
)

// BeforeCreate hooks assign IDs to persisted models before the first INSERT.
// Every hook assigns unconditionally — the backend always owns ID generation.
// Client-supplied IDs are silently overwritten; this is the security backstop.

func (m *User) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixUser)
	return nil
}

func (m *Admin) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixAdmin)
	return nil
}

func (m *ShopReferral) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixShopReferral)
	return nil
}

func (m *Role) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixRole)
	return nil
}

func (m *RolePermission) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixRolePermission)
	return nil
}

func (m *UserAddress) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixUserAddress)
	return nil
}

func (m *Address) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixAddress)
	return nil
}

func (m *Country) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixCountry)
	return nil
}

func (m *WishList) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixWishList)
	return nil
}

func (m *Cart) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixCart)
	return nil
}

func (m *CartItem) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixCartItem)
	return nil
}

func (m *Wallet) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixWallet)
	return nil
}

func (m *Transaction) BeforeCreate(*gorm.DB) error {
	m.TransactionID = NewID(PrefixTransaction)
	return nil
}

func (m *ShopVerification) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixShopVerif)
	return nil
}

func (m *ShopVerificationHistory) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixShopVerifHist)
	return nil
}

func (m *Advertisement) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixAdvertisement)
	return nil
}

func (m *SubTypeAttributes) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixSubTypeAttr)
	return nil
}

func (m *SubTypeAttributeOptions) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixSubTypeAttrOpt)
	return nil
}

func (m *CategoryImage) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixCategoryImage)
	return nil
}

func (m *PaymentMethod) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixPaymentMethod)
	return nil
}

func (m *ShopOrder) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixOrder)
	return nil
}

func (m *OrderLine) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixOrderLine)
	return nil
}

func (m *OrderReturn) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixOrderReturn)
	return nil
}

func (m *Product) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixProduct)
	return nil
}

func (m *ProductItem) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixProductItem)
	return nil
}

func (m *ProductItemImage) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixProductImage)
	return nil
}

func (m *Department) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixDepartment)
	return nil
}

func (m *Category) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixCategory)
	return nil
}

func (m *SubCategory) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixSubCategory)
	return nil
}

func (m *Brand) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixBrand)
	return nil
}

func (m *Variation) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixVariation)
	return nil
}

func (m *VariationOption) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixVariationOption)
	return nil
}

func (m *ProductImage) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixProductImage)
	return nil
}

func (m *ProductItemView) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixProductItemView)
	return nil
}

func (m *ProductItemFilterType) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixProductFilter)
	return nil
}

func (m *PromotionsType) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixPromotionType)
	return nil
}

func (m *PromotionCategory) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixPromotionCat)
	return nil
}

func (m *Promotion) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixPromotion)
	return nil
}

func (m *Offer) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixOffer)
	return nil
}

func (m *OfferCategory) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixOfferCategory)
	return nil
}

func (m *OfferProduct) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixOfferProduct)
	return nil
}

func (m *ShopDetails) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixShop)
	return nil
}

func (m *ShopOffer) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixShopOffer)
	return nil
}

func (m *ShopDepartment) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixShopDepartment)
	return nil
}

func (m *ShopTime) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixShopTime)
	return nil
}

func (m *ShopSocial) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixShopSocial)
	return nil
}

func (m *Coupon) BeforeCreate(*gorm.DB) error {
	m.CouponID = NewID(PrefixCoupon)
	return nil
}

func (m *CouponUses) BeforeCreate(*gorm.DB) error {
	m.CouponUsesID = NewID(PrefixCouponUses)
	return nil
}

func (m *Notification) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixNotification)
	return nil
}

func (m *NotificationDeviceToken) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixDeviceToken)
	return nil
}

func (m *MobileUser) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixMobileUser)
	return nil
}

func (m *OTPRequest) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixOTPRequest)
	return nil
}

func (m *LoginAuditLog) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixLoginAudit)
	return nil
}

func (m *OtpSession) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixOtpSession)
	return nil
}

func (m *OtpSessionEmail) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixOtpSession)
	return nil
}

func (m *Alert) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixAlert)
	return nil
}

func (m *AlertAction) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixAlertAction)
	return nil
}

func (m *AlertTemplate) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixAlertTemplate)
	return nil
}

func (m *SellerAlertLog) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixSellerAlertLog)
	return nil
}

func (m *Banner) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixBanner)
	return nil
}

func (m *SubscriptionPlan) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixSubscPlan)
	return nil
}

func (m *SubscriptionOrder) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixSubscOrder)
	return nil
}

func (m *UserSubscription) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixUserSubsc)
	return nil
}

func (m *FcmToken) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixDeviceToken)
	return nil
}

func (m *Job) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixJob)
	return nil
}

func (m *JobCategory) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixJobCategory)
	return nil
}

func (m *JobSubCategory) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixJobSubCategory)
	return nil
}

func (m *JobLocation) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixJobLocation)
	return nil
}

func (m *JobFilter) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixJobFilter)
	return nil
}

func (m *JobCategoryFilter) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixJobCategoryFilter)
	return nil
}

func (m *JobCategoryLocation) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixJobCategoryLocation)
	return nil
}
