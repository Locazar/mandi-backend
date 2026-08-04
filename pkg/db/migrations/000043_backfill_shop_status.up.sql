-- Backfill shop_status from the existing verification state.
--
-- Customer visibility keys off shop_status = 'active' (see shop_status_gate.go).
-- A shop is live only when an admin has verified its two mandatory checks
-- (shop photo + shop address), which is recorded as shop_verification_status =
-- TRUE. Migration 000027 was meant to do this backfill but was skipped in
-- production (historical duplicate-000027 collision), so every existing shop
-- still has shop_status = '' and is therefore invisible to customers.
--
-- Only rows whose status was never set (NULL/'') are touched, so this never
-- demotes an already-active / under_review / suspended / rejected shop. It is
-- NOT auto-activation: it activates only shops an admin already verified.
UPDATE shop_details SET shop_status = 'active'
    WHERE (shop_status IS NULL OR shop_status = '')
      AND shop_verification_status = TRUE;

UPDATE shop_details SET shop_status = 'under_review'
    WHERE (shop_status IS NULL OR shop_status = '')
      AND shop_verification_status = FALSE;
