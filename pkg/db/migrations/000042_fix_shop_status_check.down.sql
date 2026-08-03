-- This is a repair migration: its whole purpose is to make the shop_status
-- constraint correct. Reverting must NOT reintroduce the bug (a narrower CHECK
-- would reject the 'under_review'/'rejected' rows onboarding writes), so the
-- down step re-asserts the same correct constraint rather than narrowing it.
ALTER TABLE shop_details DROP CONSTRAINT IF EXISTS shop_details_shop_status_check;
ALTER TABLE shop_details ADD CONSTRAINT shop_details_shop_status_check
    CHECK (shop_status IN ('active','inactive','suspended','under_review','rejected')
        OR shop_status IS NULL OR shop_status = '');
