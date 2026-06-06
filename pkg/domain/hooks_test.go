package domain

import (
	"strings"
	"testing"

	"gorm.io/gorm"
)

// TestBeforeCreateAlwaysAssignsFreshID verifies that BeforeCreate unconditionally
// assigns a fresh prefixed ID, even when the struct already carries a value.
// This is the security backstop: the backend always owns ID generation.
// Coverage: one representative entity per domain, both empty-struct and pre-populated cases.
func TestBeforeCreateAlwaysAssignsFreshID(t *testing.T) {
	// Empty struct gets a correctly-prefixed ID.
	u := &User{}
	_ = u.BeforeCreate(nil)
	if !strings.HasPrefix(u.ID, "usr_") {
		t.Fatalf("user id = %q, want usr_ prefix", u.ID)
	}

	// Pre-populated struct is overwritten (unconditional policy).
	preset := &Admin{ID: "adm_keepme"}
	_ = preset.BeforeCreate(nil)
	if preset.ID == "adm_keepme" {
		t.Fatal("BeforeCreate must overwrite a pre-set ID (unconditional policy)")
	}
	if !strings.HasPrefix(preset.ID, "adm_") {
		t.Fatalf("admin id after BeforeCreate = %q, want adm_ prefix", preset.ID)
	}
}

// hookCase describes one entity for the table-driven overwrite test.
// getID / setID must address the actual primary-key field of the struct
// (some entities use non-standard names like CouponID or TransactionID).
type hookCase struct {
	name   string
	preset string
	run    func(preset string) (idAfter string)
	want   string // expected prefix string (without trailing "_")
}

// TestBeforeCreateOverwritesAllEntityTypes extends coverage to every domain by
// ensuring that BeforeCreate (a) assigns a non-empty ID, (b) uses the correct
// prefix, and (c) discards any pre-populated value (unconditional policy).
func TestBeforeCreateOverwritesAllEntityTypes(t *testing.T) {
	cases := []hookCase{
		// --- user domain ---
		{
			name:   "UserAddress",
			preset: "uaddr_preset",
			run:    func(p string) string { v := &UserAddress{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "uaddr",
		},
		{
			name:   "Address",
			preset: "addr_preset",
			run:    func(p string) string { v := &Address{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "addr",
		},
		{
			name:   "Country",
			preset: "ctry_preset",
			run:    func(p string) string { v := &Country{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "ctry",
		},
		{
			name:   "WishList",
			preset: "wsh_preset",
			run:    func(p string) string { v := &WishList{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "wsh",
		},
		{
			name:   "Cart",
			preset: "cart_preset",
			run:    func(p string) string { v := &Cart{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "cart",
		},
		{
			name:   "CartItem",
			preset: "citm_preset",
			run:    func(p string) string { v := &CartItem{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "citm",
		},
		{
			name:   "Wallet",
			preset: "wal_preset",
			run:    func(p string) string { v := &Wallet{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "wal",
		},
		// Transaction uses a non-standard primary key field name.
		{
			name:   "Transaction",
			preset: "txn_preset",
			run: func(p string) string {
				v := &Transaction{TransactionID: p}
				_ = v.BeforeCreate(nil)
				return v.TransactionID
			},
			want: "txn",
		},
		// --- shop domain ---
		{
			name:   "ShopDetails",
			preset: "shp_preset",
			run:    func(p string) string { v := &ShopDetails{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "shp",
		},
		{
			name:   "ShopVerification",
			preset: "shvf_preset",
			run:    func(p string) string { v := &ShopVerification{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "shvf",
		},
		{
			name:   "ShopVerificationHistory",
			preset: "shvh_preset",
			run:    func(p string) string { v := &ShopVerificationHistory{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "shvh",
		},
		{
			name:   "ShopOffer",
			preset: "shof_preset",
			run:    func(p string) string { v := &ShopOffer{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "shof",
		},
		{
			name:   "ShopDepartment",
			preset: "shdp_preset",
			run:    func(p string) string { v := &ShopDepartment{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "shdp",
		},
		{
			name:   "ShopTime",
			preset: "shtm_preset",
			run:    func(p string) string { v := &ShopTime{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "shtm",
		},
		{
			name:   "ShopSocial",
			preset: "shsoc_preset",
			run:    func(p string) string { v := &ShopSocial{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "shsoc",
		},
		// --- order domain ---
		{
			name:   "ShopOrder",
			preset: "ord_preset",
			run:    func(p string) string { v := &ShopOrder{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "ord",
		},
		{
			name:   "OrderLine",
			preset: "ordl_preset",
			run:    func(p string) string { v := &OrderLine{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "ordl",
		},
		{
			name:   "OrderReturn",
			preset: "oret_preset",
			run:    func(p string) string { v := &OrderReturn{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "oret",
		},
		{
			name:   "PaymentMethod",
			preset: "pay_preset",
			run:    func(p string) string { v := &PaymentMethod{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "pay",
		},
		// --- product domain ---
		{
			name:   "Product",
			preset: "prd_preset",
			run:    func(p string) string { v := &Product{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "prd",
		},
		{
			name:   "ProductItem",
			preset: "pri_preset",
			run:    func(p string) string { v := &ProductItem{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "pri",
		},
		{
			name:   "ProductItemImage",
			preset: "pimg_preset",
			run:    func(p string) string { v := &ProductItemImage{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "pimg",
		},
		{
			name:   "Department",
			preset: "dept_preset",
			run:    func(p string) string { v := &Department{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "dept",
		},
		{
			name:   "Category",
			preset: "cat_preset",
			run:    func(p string) string { v := &Category{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "cat",
		},
		{
			name:   "SubCategory",
			preset: "scat_preset",
			run:    func(p string) string { v := &SubCategory{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "scat",
		},
		{
			name:   "Brand",
			preset: "brnd_preset",
			run:    func(p string) string { v := &Brand{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "brnd",
		},
		{
			name:   "Variation",
			preset: "var_preset",
			run:    func(p string) string { v := &Variation{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "var",
		},
		{
			name:   "VariationOption",
			preset: "varo_preset",
			run:    func(p string) string { v := &VariationOption{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "varo",
		},
		{
			name:   "ProductImage",
			preset: "pimg_preset2",
			run:    func(p string) string { v := &ProductImage{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "pimg",
		},
		{
			name:   "ProductItemView",
			preset: "priv_preset",
			run:    func(p string) string { v := &ProductItemView{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "priv",
		},
		{
			name:   "ProductItemFilterType",
			preset: "prift_preset",
			run:    func(p string) string { v := &ProductItemFilterType{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "prift",
		},
		{
			name:   "CategoryImage",
			preset: "catim_preset",
			run:    func(p string) string { v := &CategoryImage{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "catim",
		},
		{
			name:   "SubTypeAttributes",
			preset: "sta_preset",
			run:    func(p string) string { v := &SubTypeAttributes{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "sta",
		},
		{
			name:   "SubTypeAttributeOptions",
			preset: "stao_preset",
			run:    func(p string) string { v := &SubTypeAttributeOptions{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "stao",
		},
		// --- coupon domain (non-standard PK field names) ---
		{
			name:   "Coupon",
			preset: "cpn_preset",
			run: func(p string) string {
				v := &Coupon{CouponID: p}
				_ = v.BeforeCreate(nil)
				return v.CouponID
			},
			want: "cpn",
		},
		{
			name:   "CouponUses",
			preset: "cpnu_preset",
			run: func(p string) string {
				v := &CouponUses{CouponUsesID: p}
				_ = v.BeforeCreate(nil)
				return v.CouponUsesID
			},
			want: "cpnu",
		},
		// --- offer domain ---
		{
			name:   "Offer",
			preset: "ofr_preset",
			run:    func(p string) string { v := &Offer{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "ofr",
		},
		{
			name:   "OfferCategory",
			preset: "ofrc_preset",
			run:    func(p string) string { v := &OfferCategory{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "ofrc",
		},
		{
			name:   "OfferProduct",
			preset: "ofrp_preset",
			run:    func(p string) string { v := &OfferProduct{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "ofrp",
		},
		// --- promotion domain ---
		{
			name:   "PromotionsType",
			preset: "prmt_preset",
			run:    func(p string) string { v := &PromotionsType{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "prmt",
		},
		{
			name:   "PromotionCategory",
			preset: "prmc_preset",
			run:    func(p string) string { v := &PromotionCategory{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "prmc",
		},
		{
			name:   "Promotion",
			preset: "prm_preset",
			run:    func(p string) string { v := &Promotion{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "prm",
		},
		// --- advertisement domain ---
		{
			name:   "Advertisement",
			preset: "adv_preset",
			run:    func(p string) string { v := &Advertisement{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "adv",
		},
		// --- banner domain ---
		{
			name:   "Banner",
			preset: "bnr_preset",
			run:    func(p string) string { v := &Banner{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "bnr",
		},
		// --- notification domain ---
		{
			name:   "Notification",
			preset: "ntf_preset",
			run:    func(p string) string { v := &Notification{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "ntf",
		},
		{
			name:   "NotificationDeviceToken",
			preset: "ndt_preset",
			run:    func(p string) string { v := &NotificationDeviceToken{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "ndt",
		},
		{
			name:   "FcmToken",
			preset: "ndt_preset2",
			run:    func(p string) string { v := &FcmToken{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "ndt",
		},
		// --- alert domain ---
		{
			name:   "Alert",
			preset: "alt_preset",
			run:    func(p string) string { v := &Alert{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "alt",
		},
		{
			name:   "AlertAction",
			preset: "alta_preset",
			run:    func(p string) string { v := &AlertAction{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "alta",
		},
		{
			name:   "AlertTemplate",
			preset: "altt_preset",
			run:    func(p string) string { v := &AlertTemplate{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "altt",
		},
		{
			name:   "SellerAlertLog",
			preset: "salg_preset",
			run:    func(p string) string { v := &SellerAlertLog{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "salg",
		},
		// --- subscription domain ---
		{
			name:   "SubscriptionPlan",
			preset: "subp_preset",
			run:    func(p string) string { v := &SubscriptionPlan{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "subp",
		},
		{
			name:   "SubscriptionOrder",
			preset: "subo_preset",
			run:    func(p string) string { v := &SubscriptionOrder{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "subo",
		},
		// --- auth / mobile auth domain ---
		{
			name:   "OtpSession",
			preset: "otps_preset",
			run:    func(p string) string { v := &OtpSession{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "otps",
		},
		{
			name:   "OtpSessionEmail",
			preset: "otps_preset2",
			run:    func(p string) string { v := &OtpSessionEmail{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "otps",
		},
		{
			name:   "MobileUser",
			preset: "musr_preset",
			run:    func(p string) string { v := &MobileUser{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "musr",
		},
		{
			name:   "OTPRequest",
			preset: "otpr_preset",
			run:    func(p string) string { v := &OTPRequest{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "otpr",
		},
		{
			name:   "LoginAuditLog",
			preset: "lal_preset",
			run:    func(p string) string { v := &LoginAuditLog{ID: p}; _ = v.BeforeCreate(nil); return v.ID },
			want:   "lal",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wantPrefix := tc.want + "_"
			id := tc.run(tc.preset)

			if id == "" {
				t.Fatalf("%s: BeforeCreate produced empty ID", tc.name)
			}
			if id == tc.preset {
				t.Fatalf("%s: BeforeCreate did not overwrite pre-set ID %q (unconditional policy violation)", tc.name, tc.preset)
			}
			if !strings.HasPrefix(id, wantPrefix) {
				t.Fatalf("%s: id = %q, want %s prefix", tc.name, id, wantPrefix)
			}
		})
	}
}

// TestUserSubscriptionBeforeCreateAssignsPrefixedID verifies that UserSubscription
// now uses the proper usub_ prefix instead of a raw Unix-nanosecond integer.
func TestUserSubscriptionBeforeCreateAssignsPrefixedID(t *testing.T) {
	sub := &UserSubscription{}
	_ = sub.BeforeCreate(nil)
	if !strings.HasPrefix(sub.ID, "usub_") {
		t.Fatalf("user subscription id = %q, want usub_ prefix", sub.ID)
	}

	// Pre-populated ID is also overwritten (unconditional policy).
	existing := &UserSubscription{ID: "42"}
	_ = existing.BeforeCreate(nil)
	if existing.ID == "42" {
		t.Fatal("BeforeCreate must overwrite a pre-set user subscription ID")
	}
	if !strings.HasPrefix(existing.ID, "usub_") {
		t.Fatalf("user subscription id after BeforeCreate = %q, want usub_ prefix", existing.ID)
	}
}

type jobEntity interface {
	BeforeCreate(*gorm.DB) error
}

// TestJobBeforeCreateUseDedicatedPrefixes verifies that all Job-domain hooks
// use their own dedicated prefixes rather than borrowing from other entities.
func TestJobBeforeCreateUseDedicatedPrefixes(t *testing.T) {
	type testCase struct {
		name string
		run  func() string
		want string
	}
	cases := []testCase{
		{"Job", func() string { v := &Job{}; _ = v.BeforeCreate(nil); return v.ID }, "job_"},
		{"JobCategory", func() string { v := &JobCategory{}; _ = v.BeforeCreate(nil); return v.ID }, "jcat_"},
		{"JobSubCategory", func() string { v := &JobSubCategory{}; _ = v.BeforeCreate(nil); return v.ID }, "jscat_"},
		{"JobLocation", func() string { v := &JobLocation{}; _ = v.BeforeCreate(nil); return v.ID }, "jloc_"},
		{"JobFilter", func() string { v := &JobFilter{}; _ = v.BeforeCreate(nil); return v.ID }, "jflt_"},
		{"JobCategoryFilter", func() string { v := &JobCategoryFilter{}; _ = v.BeforeCreate(nil); return v.ID }, "jcflt_"},
		{"JobCategoryLocation", func() string { v := &JobCategoryLocation{}; _ = v.BeforeCreate(nil); return v.ID }, "jcloc_"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := tc.run()
			if !strings.HasPrefix(id, tc.want) {
				t.Fatalf("%s id = %q, want %s prefix", tc.name, id, tc.want)
			}
		})
	}
}
