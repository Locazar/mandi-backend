ALTER TABLE advertisement_requests
    DROP COLUMN IF EXISTS payment_status,
    DROP COLUMN IF EXISTS payment_order_id,
    DROP COLUMN IF EXISTS payment_id,
    DROP COLUMN IF EXISTS paid_at;
