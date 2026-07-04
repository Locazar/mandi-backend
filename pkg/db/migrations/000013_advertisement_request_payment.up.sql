ALTER TABLE advertisement_requests
    ADD COLUMN IF NOT EXISTS payment_status   VARCHAR(20) NOT NULL DEFAULT 'unpaid',
    ADD COLUMN IF NOT EXISTS payment_order_id VARCHAR(64) DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS payment_id       VARCHAR(64) DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS paid_at          TIMESTAMPTZ DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_advertisement_requests_payment_order_id
    ON advertisement_requests (payment_order_id);
