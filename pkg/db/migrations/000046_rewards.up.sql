-- In-store rewards (spec: docs/superpowers/specs/2026-09-26-rewards-system-design.md).
-- Balances live in reward_accounts and are only ever changed together with an
-- append-only reward_ledger_entries row; UNIQUE(account_id, entry_type, ref_id)
-- makes every credit/debit idempotent.

CREATE TABLE IF NOT EXISTS reward_program_config (
    id                              VARCHAR(32) PRIMARY KEY,
    program_enabled                 BOOLEAN     NOT NULL DEFAULT TRUE,
    allowed_discount_percents       INT[]       NOT NULL DEFAULT '{5,10,15,20}',
    seller_base_points              BIGINT      NOT NULL DEFAULT 10,
    seller_points_per_100_rupees    BIGINT      NOT NULL DEFAULT 1,
    seller_max_points               BIGINT      NOT NULL DEFAULT 50,
    customer_base_points            BIGINT      NOT NULL DEFAULT 5,
    customer_points_per_100_rupees  BIGINT      NOT NULL DEFAULT 1,
    customer_max_points             BIGINT      NOT NULL DEFAULT 25,
    min_bill_paise                  BIGINT      NOT NULL DEFAULT 10000,
    gps_radius_m                    INT         NOT NULL DEFAULT 200,
    repeat_window_hours             INT         NOT NULL DEFAULT 168,
    claim_expiry_hours              INT         NOT NULL DEFAULT 48,
    max_claims_per_shop_per_day     INT         NOT NULL DEFAULT 30,
    max_customer_points_per_day     BIGINT      NOT NULL DEFAULT 100,
    point_value_paise               BIGINT      NOT NULL DEFAULT 100,
    seller_min_redeem_balance       BIGINT      NOT NULL DEFAULT 1000,
    customer_min_redeem_balance     BIGINT      NOT NULL DEFAULT 200,
    customer_max_redeem_pct_bp      INT         NOT NULL DEFAULT 1000,
    points_expiry_months            INT         NOT NULL DEFAULT 12,
    referral_bonus_points           BIGINT      NOT NULL DEFAULT 50,
    max_referrals_per_month         INT         NOT NULL DEFAULT 10,
    updated_by                      VARCHAR(32) NOT NULL DEFAULT '',
    updated_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO reward_program_config (id) VALUES ('rwcfg_default')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS shop_reward_settings (
    shop_id           VARCHAR(32) PRIMARY KEY REFERENCES shop_details (id) ON DELETE CASCADE,
    admin_id          VARCHAR(32) NOT NULL,
    discount_enabled  BOOLEAN     NOT NULL DEFAULT FALSE,
    discount_percent  INT         NOT NULL DEFAULT 0 CHECK (discount_percent BETWEEN 0 AND 100),
    opted_in_at       TIMESTAMPTZ,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (NOT discount_enabled OR discount_percent > 0)
);

CREATE TABLE IF NOT EXISTS reward_accounts (
    id               VARCHAR(32) PRIMARY KEY,
    owner_type       VARCHAR(16) NOT NULL CHECK (owner_type IN ('shop', 'customer')),
    owner_id         VARCHAR(32) NOT NULL,
    balance_points   BIGINT      NOT NULL DEFAULT 0 CHECK (balance_points >= 0),
    lifetime_earned  BIGINT      NOT NULL DEFAULT 0,
    lifetime_spent   BIGINT      NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_type, owner_id)
);

CREATE TABLE IF NOT EXISTS reward_ledger_entries (
    id                  VARCHAR(32) PRIMARY KEY,
    account_id          VARCHAR(32) NOT NULL REFERENCES reward_accounts (id),
    delta_points        BIGINT      NOT NULL CHECK (delta_points <> 0),
    entry_type          VARCHAR(32) NOT NULL,
    ref_type            VARCHAR(32) NOT NULL,
    ref_id              VARCHAR(64) NOT NULL,
    remaining_points    BIGINT      NOT NULL DEFAULT 0
                        CHECK (remaining_points >= 0 AND remaining_points <= GREATEST(delta_points, 0)),
    expires_at          TIMESTAMPTZ,
    expiry_reminded_at  TIMESTAMPTZ,
    note                TEXT        NOT NULL DEFAULT '',
    created_by          VARCHAR(32) NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (account_id, entry_type, ref_id)
);
CREATE INDEX IF NOT EXISTS idx_reward_ledger_account_created
    ON reward_ledger_entries (account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reward_ledger_open_lots
    ON reward_ledger_entries (expires_at) WHERE remaining_points > 0;

CREATE TABLE IF NOT EXISTS shop_purchases (
    id                           VARCHAR(32) PRIMARY KEY,
    shop_id                      VARCHAR(32) NOT NULL REFERENCES shop_details (id),
    customer_id                  VARCHAR(32) NOT NULL REFERENCES users (id),
    client_request_id            VARCHAR(64) NOT NULL,
    bill_amount_paise            BIGINT      NOT NULL CHECK (bill_amount_paise > 0),
    discount_percent             INT         NOT NULL,
    discount_paise               BIGINT      NOT NULL,
    points_redeemed              BIGINT      NOT NULL DEFAULT 0,
    points_redeemed_value_paise  BIGINT      NOT NULL DEFAULT 0,
    net_paid_paise               BIGINT      NOT NULL,
    customer_lat                 DOUBLE PRECISION NOT NULL,
    customer_lng                 DOUBLE PRECISION NOT NULL,
    distance_m                   INT         NOT NULL,
    status                       VARCHAR(16) NOT NULL DEFAULT 'pending'
                                 CHECK (status IN ('pending', 'claimed', 'rejected', 'expired')),
    seller_points                BIGINT      NOT NULL DEFAULT 0,
    customer_points              BIGINT      NOT NULL DEFAULT 0,
    expires_at                   TIMESTAMPTZ NOT NULL,
    claimed_at                   TIMESTAMPTZ,
    decided_at                   TIMESTAMPTZ,
    decided_by                   VARCHAR(32) NOT NULL DEFAULT '',
    reject_reason                TEXT        NOT NULL DEFAULT '',
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (customer_id, client_request_id)
);
CREATE INDEX IF NOT EXISTS idx_shop_purchases_pair
    ON shop_purchases (customer_id, shop_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shop_purchases_shop_status
    ON shop_purchases (shop_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shop_purchases_pending_expiry
    ON shop_purchases (expires_at) WHERE status = 'pending';
