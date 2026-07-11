-- Makes the subscription GST rate admin-configurable at runtime instead of
-- fixed via the GST_PERCENT_BASIS_POINTS env var. Seeded with that env var's
-- previous default (1800 = 18.00%) so behaviour is unchanged until an admin
-- updates it from the admin portal.

CREATE TABLE IF NOT EXISTS subscription_gst_config (
    id                     VARCHAR(32) PRIMARY KEY,
    gst_rate_basis_points  INT NOT NULL DEFAULT 1800,
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO subscription_gst_config (id, gst_rate_basis_points)
VALUES ('subgst_default', 1800)
ON CONFLICT (id) DO NOTHING;
