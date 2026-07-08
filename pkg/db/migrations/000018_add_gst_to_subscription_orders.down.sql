ALTER TABLE subscription_orders
    DROP COLUMN IF EXISTS gst_rate_basis_points,
    DROP COLUMN IF EXISTS gst_amount_amount_minor,
    DROP COLUMN IF EXISTS gst_amount_currency;
