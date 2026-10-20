DELETE FROM promotions_types WHERE id IN (
    'prmt_pct_discount','prmt_fixed_amt','prmt_flash_sale','prmt_bundle_deal',
    'prmt_bogo','prmt_free_gift','prmt_free_shipping',
    'prmt_new_customer','prmt_repeat','prmt_referral'
);
DELETE FROM promotion_categories WHERE id IN ('price_based','value_add','loyalty');
ALTER TABLE promotions_types DROP COLUMN IF EXISTS type;
