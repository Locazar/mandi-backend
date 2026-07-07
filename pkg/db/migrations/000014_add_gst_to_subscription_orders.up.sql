ALTER TABLE subscription_orders
    ADD COLUMN IF NOT EXISTS gst_rate_basis_points   INT     NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS gst_amount_amount_minor BIGINT  NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS gst_amount_currency     CHAR(3) NOT NULL DEFAULT 'INR';
