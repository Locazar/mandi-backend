package domain

import "gorm.io/gorm"

// BeforeCreate hooks assign typed-prefix IDs to every persisted model before
// the first INSERT. If a non-empty ID is already present it is kept as-is so
// that callers can supply a known ID (e.g. in tests or imports).

func (m *User) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixUser)
	}
	return nil
}

func (m *Admin) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixAdmin)
	}
	return nil
}

func (m *UserAddress) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixUserAddress)
	}
	return nil
}

func (m *Address) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixAddress)
	}
	return nil
}

func (m *Country) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixCountry)
	}
	return nil
}

func (m *WishList) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixWishList)
	}
	return nil
}

func (m *Cart) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixCart)
	}
	return nil
}

func (m *CartItem) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixCartItem)
	}
	return nil
}

func (m *Wallet) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixWallet)
	}
	return nil
}

func (m *Transaction) BeforeCreate(*gorm.DB) error {
	if m.TransactionID == "" {
		m.TransactionID = NewID(PrefixTransaction)
	}
	return nil
}

func (m *ShopVerification) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixShopVerif)
	}
	return nil
}

func (m *ShopVerificationHistory) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixShopVerifHist)
	}
	return nil
}

func (m *Advertisement) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixAdvertisement)
	}
	return nil
}

func (m *SubTypeAttributes) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixSubTypeAttr)
	}
	return nil
}

func (m *SubTypeAttributeOptions) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixSubTypeAttrOpt)
	}
	return nil
}

func (m *CategoryImage) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixCategoryImage)
	}
	return nil
}

func (m *PaymentMethod) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixPaymentMethod)
	}
	return nil
}

func (m *ShopOrder) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixOrder)
	}
	return nil
}

func (m *OrderLine) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixOrderLine)
	}
	return nil
}

func (m *OrderReturn) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixOrderReturn)
	}
	return nil
}

func (m *Product) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixProduct)
	}
	return nil
}

func (m *ProductItem) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixProductItem)
	}
	return nil
}

func (m *ProductItemImage) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixProductImage)
	}
	return nil
}

func (m *Department) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixDepartment)
	}
	return nil
}

func (m *Category) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixCategory)
	}
	return nil
}

func (m *SubCategory) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixSubCategory)
	}
	return nil
}

func (m *Brand) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixBrand)
	}
	return nil
}

func (m *Variation) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixVariation)
	}
	return nil
}

func (m *VariationOption) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixVariationOption)
	}
	return nil
}

func (m *ProductImage) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixProductImage)
	}
	return nil
}

func (m *ProductItemView) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixProductItemView)
	}
	return nil
}

func (m *ProductItemFilterType) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixProductFilter)
	}
	return nil
}

func (m *PromotionsType) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixPromotionType)
	}
	return nil
}

func (m *PromotionCategory) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixPromotionCat)
	}
	return nil
}

func (m *Promotion) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixPromotion)
	}
	return nil
}

func (m *Offer) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixOffer)
	}
	return nil
}

func (m *OfferCategory) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixOfferCategory)
	}
	return nil
}

func (m *OfferProduct) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixOfferProduct)
	}
	return nil
}

func (m *ShopDetails) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixShop)
	}
	return nil
}

func (m *ShopOffer) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixShopOffer)
	}
	return nil
}

func (m *ShopDepartment) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixShopDepartment)
	}
	return nil
}

func (m *ShopTime) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixShopTime)
	}
	return nil
}

func (m *ShopSocial) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixShopSocial)
	}
	return nil
}

func (m *Coupon) BeforeCreate(*gorm.DB) error {
	if m.CouponID == "" {
		m.CouponID = NewID(PrefixCoupon)
	}
	return nil
}

func (m *CouponUses) BeforeCreate(*gorm.DB) error {
	if m.CouponUsesID == "" {
		m.CouponUsesID = NewID(PrefixCouponUses)
	}
	return nil
}

func (m *Notification) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixNotification)
	}
	return nil
}

func (m *NotificationDeviceToken) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixDeviceToken)
	}
	return nil
}

func (m *MobileUser) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixMobileUser)
	}
	return nil
}

func (m *OTPRequest) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixOTPRequest)
	}
	return nil
}

func (m *LoginAuditLog) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixLoginAudit)
	}
	return nil
}

func (m *OtpSession) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixOtpSession)
	}
	return nil
}

func (m *OtpSessionEmail) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixOtpSession)
	}
	return nil
}

func (m *Alert) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixAlert)
	}
	return nil
}

func (m *AlertAction) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixAlertAction)
	}
	return nil
}

func (m *AlertTemplate) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixAlertTemplate)
	}
	return nil
}

func (m *SellerAlertLog) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixSellerAlertLog)
	}
	return nil
}

func (m *Banner) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixBanner)
	}
	return nil
}

func (m *SubscriptionPlan) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixSubscPlan)
	}
	return nil
}

func (m *SubscriptionOrder) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixSubscOrder)
	}
	return nil
}

func (m *UserSubscription) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixUserSubsc)
	}
	return nil
}

func (m *FcmToken) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixDeviceToken)
	}
	return nil
}

func (m *Job) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixProduct) // Jobs use product-adjacent prefix; no dedicated Job prefix in registry
	}
	return nil
}

func (m *JobCategory) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixCategory)
	}
	return nil
}

func (m *JobSubCategory) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixSubCategory)
	}
	return nil
}

func (m *JobLocation) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixAddress) // no dedicated prefix; reuse address-adjacent
	}
	return nil
}

func (m *JobFilter) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixProductFilter)
	}
	return nil
}

func (m *JobCategoryFilter) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixProductFilter)
	}
	return nil
}

func (m *JobCategoryLocation) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID(PrefixAddress)
	}
	return nil
}
