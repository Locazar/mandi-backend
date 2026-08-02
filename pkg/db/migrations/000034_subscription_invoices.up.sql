-- GST tax invoices for paid subscription orders.
--
-- subscription_invoices is INSERT-ONLY apart from the pdf_* cache columns: an
-- issued tax invoice must never be edited or deleted. Every printed value is
-- snapshotted here so a later shop rename, GST rate change, or billing-profile
-- edit cannot alter an already-issued invoice.

-- ---------------------------------------------------------------------------
-- 1. COMPANY_BILLING_PROFILE (singleton — the invoice issuer)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS company_billing_profile (
    id                    VARCHAR(32) PRIMARY KEY,
    legal_name            VARCHAR(255) NOT NULL,
    gstin                 VARCHAR(20),
    pan                   VARCHAR(20),
    address_line1         VARCHAR(255),
    address_line2         VARCHAR(255),
    city                  VARCHAR(100),
    state                 VARCHAR(100),
    state_code            VARCHAR(4),
    pincode               VARCHAR(12),
    country               VARCHAR(100),
    sac_code              VARCHAR(12),
    logo_object_key       VARCHAR(512),
    invoice_number_prefix VARCHAR(12) NOT NULL DEFAULT 'LZ',
    footer_note           TEXT,
    jurisdiction          VARCHAR(100),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- TODO: replace every placeholder below with Localzar's real registered
-- details from the admin portal BEFORE going live. SAC 998599 ("Other support
-- services n.e.c.") is a best-guess for a marketplace listing subscription —
-- confirm with a CA.
INSERT INTO company_billing_profile (
    id, legal_name, gstin, pan,
    address_line1, address_line2, city, state, state_code, pincode, country,
    sac_code, logo_object_key, invoice_number_prefix, footer_note, jurisdiction
) VALUES (
    'cbp_default',
    'Localzar Technologies Pvt. Ltd.',
    '00XXXXX0000X0Z0',
    'XXXXX0000X',
    'Address line 1', '', 'Bengaluru', 'Karnataka', '29', '560038', 'India',
    '998599',
    'locazar_icon.png',
    'LZ',
    'Prices are inclusive of GST. Computer-generated invoice — no signature required.',
    'Bengaluru'
) ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. INVOICE_NUMBER_SEQUENCES (gapless per-FY counter)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS invoice_number_sequences (
    financial_year VARCHAR(9) PRIMARY KEY,
    last_sequence  INT NOT NULL DEFAULT 0
);

-- ---------------------------------------------------------------------------
-- 3. SUBSCRIPTION_INVOICES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS subscription_invoices (
    id                      VARCHAR(32) PRIMARY KEY,
    subscription_order_id   VARCHAR(32) NOT NULL UNIQUE
        REFERENCES subscription_orders (id) ON DELETE RESTRICT,
    -- NO FK: user_id holds an admins.id for sellers, not a users.id.
    -- See 000009_relax_subscription_user_fk.
    user_id                 VARCHAR(32) NOT NULL,

    invoice_number          VARCHAR(64) NOT NULL UNIQUE,
    financial_year          VARCHAR(9)  NOT NULL,
    sequence_number         INT         NOT NULL,
    invoice_date            TIMESTAMPTZ NOT NULL,

    seller_legal_name       VARCHAR(255) NOT NULL,
    seller_gstin            VARCHAR(20),
    seller_pan              VARCHAR(20),
    seller_address          TEXT,
    seller_sac_code         VARCHAR(12),
    place_of_supply         VARCHAR(128),
    seller_footer_note      TEXT,
    seller_jurisdiction     VARCHAR(100),
    seller_logo_object_key  VARCHAR(512),

    buyer_shop_id           VARCHAR(32),
    buyer_name              VARCHAR(255),
    buyer_contact_name      VARCHAR(255),
    buyer_address           TEXT,
    buyer_gstin             VARCHAR(20),

    plan_name               VARCHAR(255),
    plan_description        TEXT,
    period_start            TIMESTAMPTZ,
    period_end              TIMESTAMPTZ,

    total_amount_minor          BIGINT  NOT NULL DEFAULT 0,
    total_currency              CHAR(3) NOT NULL DEFAULT 'INR',
    taxable_value_amount_minor  BIGINT  NOT NULL DEFAULT 0,
    taxable_value_currency      CHAR(3) NOT NULL DEFAULT 'INR',
    gst_amount_amount_minor     BIGINT  NOT NULL DEFAULT 0,
    gst_amount_currency         CHAR(3) NOT NULL DEFAULT 'INR',
    gst_rate_basis_points       INT     NOT NULL DEFAULT 0,

    razorpay_payment_id     VARCHAR(255),
    razorpay_order_id       VARCHAR(255),
    payment_method          VARCHAR(32),
    paid_at                 TIMESTAMPTZ,

    pdf_object_key          VARCHAR(512),
    pdf_generated_at        TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_subscription_invoices_fy_seq UNIQUE (financial_year, sequence_number),
    CONSTRAINT ck_subscription_invoices_total_split
        CHECK (total_amount_minor = taxable_value_amount_minor + gst_amount_amount_minor)
);

CREATE INDEX IF NOT EXISTS idx_subscription_invoices_user_id
    ON subscription_invoices (user_id);
CREATE INDEX IF NOT EXISTS idx_subscription_invoices_invoice_date
    ON subscription_invoices (invoice_date DESC);
CREATE INDEX IF NOT EXISTS idx_subscription_invoices_financial_year
    ON subscription_invoices (financial_year);
