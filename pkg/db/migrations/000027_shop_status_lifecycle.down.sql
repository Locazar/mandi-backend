-- Reverting is only safe if no rows currently use 'under_review'/'rejected' —
-- this restores the original 3-value constraint and will fail (by design)
-- if such rows still exist, since they'd violate it.
ALTER TABLE shop_details DROP CONSTRAINT IF EXISTS shop_details_shop_status_check;
ALTER TABLE shop_details ADD CONSTRAINT shop_details_shop_status_check
    CHECK (shop_status IN ('active','inactive','suspended')
        OR shop_status IS NULL OR shop_status = '');
