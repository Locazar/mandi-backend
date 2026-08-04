-- Pure data backfill. Reverting would re-blank shop_status and re-hide every
-- shop from customers, which is never desirable — intentional no-op.
SELECT 1;
