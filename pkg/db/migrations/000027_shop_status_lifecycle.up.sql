-- Extends shop_status into a real approval lifecycle: a newly-onboarded shop
-- now enters 'under_review' (invisible to customers) until an admin approves
-- it to 'active', or rejects it to 'rejected' (seller can fix & resubmit,
-- which returns it to 'under_review'). 'suspended'/'inactive' are unchanged.
ALTER TABLE shop_details DROP CONSTRAINT IF EXISTS shop_details_shop_status_check;
ALTER TABLE shop_details ADD CONSTRAINT shop_details_shop_status_check
    CHECK (shop_status IN ('active','inactive','suspended','under_review','rejected')
        OR shop_status IS NULL OR shop_status = '');

-- Backfill only rows that have never had a status set (shop_status was
-- previously written just once, during onboarding, and only when the
-- request included it). Rows already carrying an explicit status
-- (active/inactive/suspended) are left untouched.
UPDATE shop_details
SET shop_status = 'active'
WHERE (shop_status IS NULL OR shop_status = '')
  AND shop_verification_status = TRUE;

UPDATE shop_details
SET shop_status = 'under_review'
WHERE (shop_status IS NULL OR shop_status = '')
  AND shop_verification_status = FALSE;
