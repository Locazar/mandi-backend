package domain

import (
	"strings"

	"github.com/nrednav/cuid2"
)

// IDPrefix is the per-entity prefix component of a typed-prefix identifier.
// Every persisted entity uses an opaque, non-enumerable string PK of the form
// "<prefix>_<cuid2>" (e.g. "usr_x7k2q9m4p3..."), generated application-side.
type IDPrefix string

const idSeparator = "_"

const (
	PrefixUser            IDPrefix = "usr"
	PrefixAddress         IDPrefix = "addr"
	PrefixUserAddress     IDPrefix = "uaddr"
	PrefixCountry         IDPrefix = "ctry"
	PrefixAdmin           IDPrefix = "adm"
	PrefixShopVerif       IDPrefix = "shvf"
	PrefixShopVerifHist   IDPrefix = "shvh"
	PrefixAgent           IDPrefix = "agt"
	PrefixCategory        IDPrefix = "cat"
	PrefixCategoryImage   IDPrefix = "catim"
	PrefixSubCategory     IDPrefix = "scat"
	PrefixSubCategoryDet  IDPrefix = "scatd"
	PrefixDepartment      IDPrefix = "dept"
	PrefixProduct         IDPrefix = "prd"
	PrefixProductItem     IDPrefix = "pri"
	PrefixProductItemView IDPrefix = "priv"
	PrefixProductImage    IDPrefix = "pimg"
	PrefixProductConfig   IDPrefix = "prcfg"
	PrefixProductFilter   IDPrefix = "prift"
	PrefixBrand           IDPrefix = "brnd"
	PrefixVariation       IDPrefix = "var"
	PrefixVariationOption IDPrefix = "varo"
	PrefixSubTypeAttr     IDPrefix = "sta"
	PrefixSubTypeAttrOpt  IDPrefix = "stao"
	PrefixWishList        IDPrefix = "wsh"
	PrefixCart            IDPrefix = "cart"
	PrefixCartItem        IDPrefix = "citm"
	PrefixOrder           IDPrefix = "ord"
	PrefixOrderLine       IDPrefix = "ordl"
	PrefixOrderReturn     IDPrefix = "oret"
	PrefixFeedback        IDPrefix = "fbk"
	PrefixOffer           IDPrefix = "ofr"
	PrefixOfferCategory   IDPrefix = "ofrc"
	PrefixOfferProduct    IDPrefix = "ofrp"
	PrefixCoupon          IDPrefix = "cpn"
	PrefixCouponUses      IDPrefix = "cpnu"
	PrefixWallet          IDPrefix = "wal"
	PrefixTransaction     IDPrefix = "txn"
	PrefixAdvertisement   IDPrefix = "adv"
	PrefixNotification    IDPrefix = "ntf"
	PrefixDeviceToken     IDPrefix = "ndt"
	PrefixShop            IDPrefix = "shp"
	PrefixShopOffer       IDPrefix = "shof"
	PrefixShopDepartment  IDPrefix = "shdp"
	PrefixShopTime        IDPrefix = "shtm"
	PrefixShopSocial      IDPrefix = "shsoc"
	PrefixPaymentMethod   IDPrefix = "pay"
	PrefixAlert           IDPrefix = "alt"
	PrefixAlertAction     IDPrefix = "alta"
	PrefixAlertTemplate   IDPrefix = "altt"
	PrefixSellerAlertLog  IDPrefix = "salg"
	PrefixPromotion       IDPrefix = "prm"
	PrefixPromotionType   IDPrefix = "prmt"
	PrefixPromotionCat    IDPrefix = "prmc"
	PrefixBanner          IDPrefix = "bnr"
	PrefixSubscPlan       IDPrefix = "subp"
	PrefixSubscOrder      IDPrefix = "subo"
	PrefixUserSubsc       IDPrefix = "usub"
	PrefixOTPRequest      IDPrefix = "otpr"
	PrefixLoginAudit      IDPrefix = "lal"
	PrefixUserConsent     IDPrefix = "cnst"
	PrefixAuditLog        IDPrefix = "aud"
	PrefixUserOfferHist   IDPrefix = "uoh"
)

// allPrefixes is the authoritative registry of every entity prefix. Used by
// tests to guarantee uniqueness and bounded length.
var allPrefixes = []IDPrefix{
	PrefixUser, PrefixAddress, PrefixUserAddress, PrefixCountry, PrefixAdmin,
	PrefixShopVerif, PrefixShopVerifHist, PrefixAgent, PrefixCategory,
	PrefixCategoryImage, PrefixSubCategory, PrefixSubCategoryDet, PrefixDepartment,
	PrefixProduct, PrefixProductItem, PrefixProductItemView, PrefixProductImage,
	PrefixProductConfig, PrefixProductFilter, PrefixBrand, PrefixVariation,
	PrefixVariationOption, PrefixSubTypeAttr, PrefixSubTypeAttrOpt, PrefixWishList,
	PrefixCart, PrefixCartItem, PrefixOrder, PrefixOrderLine, PrefixOrderReturn,
	PrefixFeedback, PrefixOffer, PrefixOfferCategory, PrefixOfferProduct,
	PrefixCoupon, PrefixCouponUses, PrefixWallet, PrefixTransaction,
	PrefixAdvertisement, PrefixNotification, PrefixDeviceToken, PrefixShop,
	PrefixShopOffer, PrefixShopDepartment, PrefixShopTime, PrefixShopSocial,
	PrefixPaymentMethod, PrefixAlert, PrefixAlertAction, PrefixAlertTemplate,
	PrefixSellerAlertLog, PrefixPromotion, PrefixPromotionType, PrefixPromotionCat,
	PrefixBanner, PrefixSubscPlan, PrefixSubscOrder, PrefixUserSubsc,
	PrefixOTPRequest, PrefixLoginAudit, PrefixUserConsent, PrefixAuditLog,
	PrefixUserOfferHist,
}

// NewID returns a fresh typed-prefix identifier, e.g. "usr_x7k2q9m4p3...".
func NewID(p IDPrefix) string {
	return string(p) + idSeparator + cuid2.Generate()
}

// PrefixOf extracts the prefix portion of an id, or "" if it is unprefixed.
func PrefixOf(id string) IDPrefix {
	if i := strings.Index(id, idSeparator); i > 0 {
		return IDPrefix(id[:i])
	}
	return ""
}

// HasPrefix reports whether id carries the given entity prefix.
func HasPrefix(id string, p IDPrefix) bool {
	return strings.HasPrefix(id, string(p)+idSeparator)
}
