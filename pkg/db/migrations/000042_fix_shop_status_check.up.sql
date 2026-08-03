-- Repair: re-assert the shop_status lifecycle CHECK constraint.
--
-- Migration 000027_shop_status_lifecycle widened this constraint to allow
-- 'under_review'/'rejected', but a historical duplicate-000027 collision left
-- production with version 27 recorded as applied while that ALTER never
-- actually ran. The old baseline constraint (active/inactive/suspended only)
-- therefore persisted. Now that onboarding defaults a new shop to
-- 'under_review' (COALESCE(..., 'under_review') in the shop upsert), every
-- shop INSERT fails the stale CHECK and returns 500.
--
-- This is idempotent and non-destructive: DROP IF EXISTS + ADD, no data is
-- mutated, and *widening* a CHECK can never reject existing rows or existing
-- writes. A fresh version number guarantees migrate runs it (unlike 27).
ALTER TABLE shop_details DROP CONSTRAINT IF EXISTS shop_details_shop_status_check;
ALTER TABLE shop_details ADD CONSTRAINT shop_details_shop_status_check
    CHECK (shop_status IN ('active','inactive','suspended','under_review','rejected')
        OR shop_status IS NULL OR shop_status = '');
