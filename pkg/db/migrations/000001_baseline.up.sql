-- Drop existing triggers to allow schema changes
DROP TRIGGER IF EXISTS update_cart_total_price ON cart_items CASCADE;
DROP TRIGGER IF EXISTS update_product_quantity ON order_lines CASCADE;
DROP TRIGGER IF EXISTS update_product_qty_on_order_return ON shop_orders CASCADE;
DROP FUNCTION IF EXISTS update_cart_total_price() CASCADE;
DROP FUNCTION IF EXISTS update_product_quantity() CASCADE;
DROP FUNCTION IF EXISTS update_product_quantity_on_return() CASCADE;

-- 000001_baseline.up.sql
-- Full corrected end-state schema for mandi-backend.
-- String PKs (typed-prefix VARCHAR(32)), Money columns, CHECK enums,
-- FK constraints, soft-delete, indexes, DPDP tables.
-- Table ordering: FK targets before referencers.

-- ---------------------------------------------------------------------------
-- 1. COUNTRIES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS countries (
    id          VARCHAR(32) PRIMARY KEY,
    country_name VARCHAR(100) NOT NULL,
    iso_code    VARCHAR(10)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_countries_country_name ON countries (country_name);
CREATE UNIQUE INDEX IF NOT EXISTS uidx_countries_iso_code     ON countries (iso_code);

-- ---------------------------------------------------------------------------
-- 2. USERS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id              VARCHAR(32) PRIMARY KEY,
    date_of_birth   DATE,
    first_name      VARCHAR(255),
    last_name       VARCHAR(255),
    email           VARCHAR(255),
    phone           VARCHAR(50) NOT NULL,
    password        TEXT        NOT NULL,
    email_verified  BOOLEAN     NOT NULL DEFAULT FALSE,
    phone_verified  BOOLEAN     NOT NULL DEFAULT FALSE,
    block_status    BOOLEAN     NOT NULL DEFAULT FALSE,
    trial_used      BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_users_phone      ON users (phone);
CREATE UNIQUE INDEX IF NOT EXISTS uidx_users_email      ON users (email) WHERE email <> '';
CREATE INDEX        IF NOT EXISTS idx_users_deleted_at   ON users (deleted_at);

-- ---------------------------------------------------------------------------
-- 3. ADMINS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS admins (
    id                   VARCHAR(32) PRIMARY KEY,
    user_name            VARCHAR(255),
    full_name            VARCHAR(255),
    email                VARCHAR(255),
    password             TEXT,
    address_line1        VARCHAR(255),
    address_line2        VARCHAR(255),
    city                 VARCHAR(50),
    state                VARCHAR(50),
    country              VARCHAR(50),
    pincode              VARCHAR(50),
    mobile               VARCHAR(50),
    profile_image_url    VARCHAR(255),
    latitude             DECIMAL(10,7),
    longitude            DECIMAL(10,7),
    payment_status       BOOLEAN     NOT NULL DEFAULT FALSE,
    payment_type         VARCHAR(50),
    payment_date         TIMESTAMPTZ,
    start_date           TIMESTAMPTZ,
    expiry_date          TIMESTAMPTZ,
    bank_account_number  TEXT,
    bank_ifsc            TEXT,
    pan                  TEXT,
    aadhaar_last4        VARCHAR(4),
    aadhaar_verified     BOOLEAN     NOT NULL DEFAULT FALSE,
    verified_seller      BOOLEAN     NOT NULL DEFAULT FALSE,
    status               VARCHAR(50)
        CHECK (status IN ('active','inactive','suspended')),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_admins_deleted_at ON admins (deleted_at);

-- ---------------------------------------------------------------------------
-- 4. ADDRESSES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS addresses (
    id            VARCHAR(32) PRIMARY KEY,
    user_id       VARCHAR(32) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    land_mark     VARCHAR(255) NOT NULL,
    area          VARCHAR(255),
    city          VARCHAR(255) NOT NULL,
    pincode       VARCHAR(10),
    country_id    VARCHAR(32) NOT NULL REFERENCES countries (id) ON DELETE RESTRICT,
    latitude      DECIMAL(10,7),
    longitude     DECIMAL(10,7),
    phone_number  VARCHAR(255) NOT NULL,
    address_type  VARCHAR(50) NOT NULL
        CHECK (address_type IN ('home','work','other')),
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    is_default    BOOLEAN,
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ  NULL
);

CREATE INDEX IF NOT EXISTS idx_addresses_user_id    ON addresses (user_id);
CREATE INDEX IF NOT EXISTS idx_addresses_country_id ON addresses (country_id);
CREATE INDEX IF NOT EXISTS idx_addresses_deleted_at ON addresses (deleted_at);

-- ---------------------------------------------------------------------------
-- 5. USER_ADDRESSES (join table)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_addresses (
    id         VARCHAR(32) PRIMARY KEY,
    user_id    VARCHAR(32) NOT NULL REFERENCES users    (id) ON DELETE CASCADE,
    address_id VARCHAR(32) NOT NULL REFERENCES addresses (id) ON DELETE CASCADE,
    is_default BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_addresses_user_id    ON user_addresses (user_id);
CREATE INDEX IF NOT EXISTS idx_user_addresses_address_id ON user_addresses (address_id);

-- ---------------------------------------------------------------------------
-- 6. PAYMENT_METHODS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS payment_methods (
    id                         VARCHAR(32) PRIMARY KEY,
    name                       VARCHAR(50) NOT NULL
        CHECK (name IN ('razor pay','cod','stripe')),
    block_status               BOOLEAN     NOT NULL DEFAULT FALSE,
    maximum_amount_amount_minor BIGINT      NOT NULL DEFAULT 0,
    maximum_amount_currency     CHAR(3)     NOT NULL DEFAULT 'INR'
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_payment_methods_name ON payment_methods (name);

-- ---------------------------------------------------------------------------
-- 7. SHOP_DETAILS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shop_details (
    id                          VARCHAR(32) PRIMARY KEY,
    admin_id                    VARCHAR(32) NOT NULL REFERENCES admins (id) ON DELETE CASCADE,
    shop_name                   VARCHAR(100),
    owner_name                  VARCHAR(100),
    email                       VARCHAR(100),
    phone                       VARCHAR(50),
    address_line1               VARCHAR(255),
    address_line2               VARCHAR(255),
    city                        VARCHAR(50),
    state                       VARCHAR(50),
    country                     VARCHAR(50),
    pincode                     VARCHAR(50),
    latitude                    DECIMAL(10,7),
    longitude                   DECIMAL(10,7),
    shop_description            TEXT,
    shop_verification_docs      TEXT,
    document_type               VARCHAR(50)
        CHECK (document_type IN ('gst','pan','aadhaar','license') OR document_type IS NULL OR document_type = ''),
    document_value              TEXT,
    pan_number                  TEXT,
    itr_documents               TEXT,
    shop_type                   VARCHAR(50)
        CHECK (shop_type IN ('retail','wholesale','service') OR shop_type IS NULL OR shop_type = ''),
    shop_status                 VARCHAR(50)
        CHECK (shop_status IN ('active','inactive','suspended') OR shop_status IS NULL OR shop_status = ''),
    bank_account_number         TEXT,
    bank_ifsc                   TEXT,
    shop_image_url              VARCHAR(255),
    shop_verification_status        BOOLEAN NOT NULL DEFAULT FALSE,
    shop_verification_remarks       TEXT    NOT NULL DEFAULT '',
    photo_shop_verification         BOOLEAN NOT NULL DEFAULT FALSE,
    business_doc_verification       BOOLEAN NOT NULL DEFAULT FALSE,
    identity_doc_verification       BOOLEAN NOT NULL DEFAULT FALSE,
    address_proof_verification      BOOLEAN NOT NULL DEFAULT FALSE,
    has_offers                  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_shop_details_admin_id ON shop_details (admin_id);
CREATE INDEX        IF NOT EXISTS idx_shop_details_deleted_at ON shop_details (deleted_at);

-- ---------------------------------------------------------------------------
-- 8. DEPARTMENTS, CATEGORIES, SUB_CATEGORIES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS departments (
    id         VARCHAR(32) PRIMARY KEY,
    name       VARCHAR(50) NOT NULL UNIQUE,
    sort_order INT         NOT NULL DEFAULT 0,
    is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
    image_url  TEXT,
    icon       VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS categories (
    id            VARCHAR(32) PRIMARY KEY,
    department_id VARCHAR(32) NOT NULL REFERENCES departments (id) ON DELETE RESTRICT,
    name          VARCHAR(30) NOT NULL,
    sort_order    INT         NOT NULL DEFAULT 0,
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    image_url     TEXT,
    icon          VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_categories_department_id ON categories (department_id);

CREATE TABLE IF NOT EXISTS sub_categories (
    id            VARCHAR(32) PRIMARY KEY,
    department_id VARCHAR(32) NOT NULL REFERENCES departments  (id) ON DELETE RESTRICT,
    category_id   VARCHAR(32) NOT NULL REFERENCES categories  (id) ON DELETE RESTRICT,
    name          VARCHAR(30) NOT NULL,
    sort_order    INT         NOT NULL DEFAULT 0,
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    image_url     TEXT,
    icon          VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_sub_categories_department_id ON sub_categories (department_id);
CREATE INDEX IF NOT EXISTS idx_sub_categories_category_id   ON sub_categories (category_id);

-- ---------------------------------------------------------------------------
-- 9. BRANDS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS brands (
    id         VARCHAR(32) PRIMARY KEY,
    name       VARCHAR(255) NOT NULL UNIQUE,
    sort_order INT          NOT NULL DEFAULT 0,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE
);

-- ---------------------------------------------------------------------------
-- 10. VARIATIONS & VARIATION_OPTIONS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS variations (
    id              VARCHAR(32) PRIMARY KEY,
    sub_category_id VARCHAR(32) NOT NULL REFERENCES sub_categories (id) ON DELETE RESTRICT,
    name            VARCHAR(255) NOT NULL,
    sort_order      INT          NOT NULL DEFAULT 0,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_variations_sub_category_id ON variations (sub_category_id);

CREATE TABLE IF NOT EXISTS variation_options (
    id           VARCHAR(32) PRIMARY KEY,
    variation_id VARCHAR(32) NOT NULL REFERENCES variations (id) ON DELETE RESTRICT,
    value        VARCHAR(255) NOT NULL,
    sort_order   INT          NOT NULL DEFAULT 0,
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_variation_options_variation_id ON variation_options (variation_id);

-- ---------------------------------------------------------------------------
-- 11. SUB_TYPE_ATTRIBUTES & SUB_TYPE_ATTRIBUTE_OPTIONS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sub_type_attributes (
    id              VARCHAR(32) PRIMARY KEY,
    sub_category_id VARCHAR(32) NOT NULL REFERENCES sub_categories (id) ON DELETE CASCADE,
    field_name      VARCHAR(50) NOT NULL,
    field_type      VARCHAR(20) NOT NULL
        CHECK (field_type IN ('dropdown','number','text')),
    is_required     BOOLEAN     NOT NULL DEFAULT TRUE,
    sort_order      INT         NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_sub_type_attributes_sub_category_id ON sub_type_attributes (sub_category_id);

CREATE TABLE IF NOT EXISTS sub_type_attribute_options (
    id                    VARCHAR(32) PRIMARY KEY,
    sub_type_attribute_id VARCHAR(32) NOT NULL REFERENCES sub_type_attributes (id) ON DELETE CASCADE,
    option_value          VARCHAR(50) NOT NULL,
    sort_order            INT         NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_sub_type_attr_opts_attr_id ON sub_type_attribute_options (sub_type_attribute_id);

-- ---------------------------------------------------------------------------
-- 12. PRODUCTS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS products (
    id            VARCHAR(32) PRIMARY KEY,
    name          VARCHAR(50) NOT NULL,
    description   VARCHAR(100) NOT NULL,
    category_id   VARCHAR(32) REFERENCES categories (id) ON DELETE SET NULL,
    department_id VARCHAR(32) REFERENCES departments (id) ON DELETE SET NULL,
    image         TEXT        NOT NULL,
    shop_id       VARCHAR(32) NOT NULL REFERENCES shop_details (id) ON DELETE CASCADE,
    created_by    VARCHAR(32),
    updated_by    VARCHAR(32),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_shop_id     ON products (shop_id);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products (category_id);
CREATE INDEX IF NOT EXISTS idx_products_shop_active ON products (shop_id, created_at);

-- ---------------------------------------------------------------------------
-- 13. PRODUCT_ITEMS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS product_items (
    id                   VARCHAR(32) PRIMARY KEY,
    sub_category_name    VARCHAR(255) NOT NULL,
    sub_category_id      VARCHAR(32) REFERENCES sub_categories (id) ON DELETE SET NULL,
    category_id          VARCHAR(32) REFERENCES categories     (id) ON DELETE SET NULL,
    department_id        VARCHAR(32) REFERENCES departments    (id) ON DELETE SET NULL,
    dynamic_fields       JSONB        NOT NULL,
    admin_id             VARCHAR(32)  REFERENCES admins        (id) ON DELETE SET NULL,
    product_item_images  TEXT[],
    shop_id              VARCHAR(32)  REFERENCES shop_details  (id) ON DELETE CASCADE,
    created_by           VARCHAR(32),
    updated_by           VARCHAR(32),
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    stock                BOOLEAN      NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_product_items_shop_id     ON product_items (shop_id);
CREATE INDEX IF NOT EXISTS idx_product_items_admin_id    ON product_items (admin_id);
CREATE INDEX IF NOT EXISTS idx_product_items_shop_active ON product_items (shop_id, stock);

-- ---------------------------------------------------------------------------
-- 14. PRODUCT_IMAGES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS product_images (
    id              VARCHAR(32) PRIMARY KEY,
    product_item_id VARCHAR(32) NOT NULL REFERENCES product_items (id) ON DELETE CASCADE,
    image_url       TEXT[],
    shop_id         VARCHAR(32) NOT NULL REFERENCES shop_details  (id) ON DELETE CASCADE,
    product_id      VARCHAR(32) NOT NULL REFERENCES products      (id) ON DELETE CASCADE,
    alt_text        VARCHAR(255),
    sort_order      INT         NOT NULL DEFAULT 0,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_images_product_item_id ON product_images (product_item_id);
CREATE INDEX IF NOT EXISTS idx_product_images_product_id      ON product_images (product_id);

-- ---------------------------------------------------------------------------
-- 15. PRODUCT_ITEM_IMAGES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS product_item_images (
    id              VARCHAR(32) PRIMARY KEY,
    product_item_id VARCHAR(32) NOT NULL REFERENCES product_items (id) ON DELETE CASCADE,
    image_urls      TEXT[],
    alt_text        VARCHAR(255),
    sort_order      INT         NOT NULL DEFAULT 0,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    shop_id         VARCHAR(32) NOT NULL REFERENCES shop_details  (id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_item_images_product_item_id ON product_item_images (product_item_id);

-- ---------------------------------------------------------------------------
-- 16. PRODUCT_CONFIGURATIONS (composite-key join)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS product_configurations (
    product_item_id     VARCHAR(32) NOT NULL REFERENCES product_items     (id) ON DELETE CASCADE,
    variation_option_id VARCHAR(32) NOT NULL REFERENCES variation_options (id) ON DELETE RESTRICT,
    sort_order          INT         NOT NULL DEFAULT 0,
    is_active           BOOLEAN     NOT NULL DEFAULT TRUE,
    PRIMARY KEY (product_item_id, variation_option_id)
);

CREATE INDEX IF NOT EXISTS idx_product_configurations_variation_option_id ON product_configurations (variation_option_id);

-- ---------------------------------------------------------------------------
-- 17. PRODUCT_ITEM_VIEWS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS product_item_views (
    id              VARCHAR(32) PRIMARY KEY,
    product_item_id VARCHAR(32) NOT NULL REFERENCES product_items (id) ON DELETE CASCADE,
    shop_id         VARCHAR(32) NOT NULL REFERENCES shop_details  (id) ON DELETE CASCADE,
    admin_id        VARCHAR(32),
    viewed_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    view_count      INT         NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_item_views_product_item_id ON product_item_views (product_item_id);
CREATE INDEX IF NOT EXISTS idx_product_item_views_admin_id        ON product_item_views (admin_id);

-- ---------------------------------------------------------------------------
-- 18. PRODUCT_ITEM_FILTER_TYPES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS product_item_filter_types (
    id          VARCHAR(32) PRIMARY KEY,
    filter_name VARCHAR(255) NOT NULL,
    shop_id     VARCHAR(32)  REFERENCES shop_details (id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_item_filter_types_shop_id ON product_item_filter_types (shop_id);

-- ---------------------------------------------------------------------------
-- 19. CATEGORY_IMAGES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS category_images (
    id          VARCHAR(32) PRIMARY KEY,
    category_id VARCHAR(32) NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    image_url   TEXT        NOT NULL,
    alt_text    VARCHAR(255),
    sort_order  INT         NOT NULL DEFAULT 0,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_category_images_category_id ON category_images (category_id);

-- ---------------------------------------------------------------------------
-- 20. OFFERS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS offers (
    id            VARCHAR(32) PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    description   VARCHAR(255) NOT NULL,
    discount_rate BIGINT       NOT NULL
        CHECK (discount_rate BETWEEN 0 AND 100),
    offer_type    VARCHAR(50)  NOT NULL
        CHECK (offer_type IN ('percentage','fixed')),
    start_date    TIMESTAMPTZ  NOT NULL,
    end_date      TIMESTAMPTZ  NOT NULL,
    image         TEXT         NOT NULL,
    thumbnail     TEXT,
    created_by    VARCHAR(32),
    updated_by    VARCHAR(32),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    sort_order    INT          NOT NULL DEFAULT 0,
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE
);

-- ---------------------------------------------------------------------------
-- 21. OFFER_CATEGORIES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS offer_categories (
    id          VARCHAR(32) PRIMARY KEY,
    offer_id    VARCHAR(32) NOT NULL REFERENCES offers     (id) ON DELETE CASCADE,
    category_id VARCHAR(32) NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    sort_order  INT         NOT NULL DEFAULT 0,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_offer_categories_offer_id    ON offer_categories (offer_id);
CREATE INDEX IF NOT EXISTS idx_offer_categories_category_id ON offer_categories (category_id);

-- ---------------------------------------------------------------------------
-- 22. OFFER_PRODUCTS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS offer_products (
    id              VARCHAR(32) PRIMARY KEY,
    offer_id        VARCHAR(32) NOT NULL REFERENCES offers        (id) ON DELETE CASCADE,
    product_item_id VARCHAR(32) NOT NULL REFERENCES product_items (id) ON DELETE CASCADE,
    sort_order      INT         NOT NULL DEFAULT 0,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_offer_products_offer_id        ON offer_products (offer_id);
CREATE INDEX IF NOT EXISTS idx_offer_products_product_item_id ON offer_products (product_item_id);

-- ---------------------------------------------------------------------------
-- 23. COUPONS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS coupons (
    coupon_id                     VARCHAR(32) PRIMARY KEY,
    coupon_name                   VARCHAR(25)  NOT NULL UNIQUE,
    coupon_code                   VARCHAR(255) NOT NULL UNIQUE,
    expire_date                   TIMESTAMPTZ  NOT NULL,
    description                   VARCHAR(255) NOT NULL,
    discount_rate                 BIGINT       NOT NULL
        CHECK (discount_rate BETWEEN 0 AND 100),
    minimum_cart_price_amount_minor BIGINT       NOT NULL DEFAULT 0,
    minimum_cart_price_currency     CHAR(3)      NOT NULL DEFAULT 'INR',
    image                         TEXT,
    block_status                  BOOLEAN      NOT NULL DEFAULT FALSE,
    created_by                    VARCHAR(32),
    updated_by                    VARCHAR(32),
    created_at                    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at                    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ---------------------------------------------------------------------------
-- 24. COUPON_USES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS coupon_uses (
    coupon_uses_id VARCHAR(32) PRIMARY KEY,
    coupon_id      VARCHAR(32) NOT NULL REFERENCES coupons (coupon_id) ON DELETE RESTRICT,
    user_id        VARCHAR(32) NOT NULL REFERENCES users   (id)        ON DELETE RESTRICT,
    used_at        TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_coupon_uses_coupon_id ON coupon_uses (coupon_id);
CREATE INDEX IF NOT EXISTS idx_coupon_uses_user_id   ON coupon_uses (user_id);

-- ---------------------------------------------------------------------------
-- 25. WALLETS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS wallets (
    id                       VARCHAR(32) PRIMARY KEY,
    user_id                  VARCHAR(32) NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    total_amount_amount_minor BIGINT      NOT NULL DEFAULT 0,
    total_amount_currency     CHAR(3)     NOT NULL DEFAULT 'INR',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets (user_id);

-- ---------------------------------------------------------------------------
-- 26. TRANSACTIONS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS transactions (
    transaction_id          VARCHAR(32) PRIMARY KEY,
    wallet_id               VARCHAR(32) NOT NULL REFERENCES wallets (id) ON DELETE RESTRICT,
    transaction_date        TIMESTAMPTZ NOT NULL,
    amount_amount_minor      BIGINT      NOT NULL DEFAULT 0,
    amount_currency          CHAR(3)     NOT NULL DEFAULT 'INR',
    transaction_type        VARCHAR(20) NOT NULL
        CHECK (transaction_type IN ('DEBIT','CREDIT'))
);

CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id ON transactions (wallet_id);

-- ---------------------------------------------------------------------------
-- 27. CARTS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS carts (
    id                              VARCHAR(32) PRIMARY KEY,
    user_id                         VARCHAR(32) NOT NULL REFERENCES users   (id)   ON DELETE CASCADE,
    total_price_amount_minor         BIGINT      NOT NULL DEFAULT 0,
    total_price_currency             CHAR(3)     NOT NULL DEFAULT 'INR',
    applied_coupon_id               VARCHAR(32) REFERENCES coupons (coupon_id) ON DELETE SET NULL,
    discount_amount_amount_minor     BIGINT      NOT NULL DEFAULT 0,
    discount_amount_currency         CHAR(3)     NOT NULL DEFAULT 'INR',
    created_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_carts_user_id   ON carts (user_id);

-- ---------------------------------------------------------------------------
-- 28. CART_ITEMS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS cart_items (
    id              VARCHAR(32) PRIMARY KEY,
    cart_id         VARCHAR(32) NOT NULL REFERENCES carts        (id)  ON DELETE CASCADE,
    product_item_id VARCHAR(32) NOT NULL REFERENCES product_items (id) ON DELETE CASCADE,
    qty             INT         NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id         ON cart_items (cart_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_product_item_id ON cart_items (product_item_id);
CREATE UNIQUE INDEX IF NOT EXISTS uidx_cart_items_cart_product ON cart_items (cart_id, product_item_id);

-- ---------------------------------------------------------------------------
-- 29. WISH_LISTS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS wish_lists (
    id              VARCHAR(32) PRIMARY KEY,
    user_id         VARCHAR(32) NOT NULL REFERENCES users        (id)  ON DELETE CASCADE,
    shop_id         VARCHAR(32) NOT NULL REFERENCES shop_details (id)  ON DELETE CASCADE,
    admin_id        VARCHAR(32),
    product_item_id VARCHAR(32) NOT NULL REFERENCES product_items (id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wish_lists_user_id         ON wish_lists (user_id);
CREATE INDEX IF NOT EXISTS idx_wish_lists_product_item_id ON wish_lists (product_item_id);
CREATE UNIQUE INDEX IF NOT EXISTS uidx_wish_lists_user_shop_product ON wish_lists (user_id, product_item_id, shop_id);

-- ---------------------------------------------------------------------------
-- 30. SHOP_ORDERS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shop_orders (
    id                              VARCHAR(32) PRIMARY KEY,
    user_id                         VARCHAR(32) NOT NULL REFERENCES users          (id) ON DELETE RESTRICT,
    order_date                      TIMESTAMPTZ NOT NULL,
    address_id                      VARCHAR(32) NOT NULL REFERENCES addresses      (id) ON DELETE RESTRICT,
    order_total_amount_minor         BIGINT      NOT NULL DEFAULT 0,
    order_total_currency             CHAR(3)     NOT NULL DEFAULT 'INR',
    discount_amount_minor            BIGINT      NOT NULL DEFAULT 0,
    discount_currency                CHAR(3)     NOT NULL DEFAULT 'INR',
    status                          VARCHAR(50) NOT NULL
        CHECK (status IN (
            'payment pending','order placed','order cancelled','order delivered',
            'return requested','return approved','return cancelled','order returned'
        )),
    payment_method_id               VARCHAR(32) REFERENCES payment_methods (id) ON DELETE SET NULL,
    shop_id                         VARCHAR(32) REFERENCES shop_details    (id) ON DELETE SET NULL,
    deleted_at                      TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_shop_orders_user_id          ON shop_orders (user_id);
CREATE INDEX IF NOT EXISTS idx_shop_orders_shop_id          ON shop_orders (shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_orders_payment_method_id ON shop_orders (payment_method_id);
CREATE INDEX IF NOT EXISTS idx_shop_orders_deleted_at        ON shop_orders (deleted_at);
CREATE INDEX IF NOT EXISTS idx_shop_orders_status            ON shop_orders (status);

-- ---------------------------------------------------------------------------
-- 31. ORDER_LINES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS order_lines (
    id              VARCHAR(32) PRIMARY KEY,
    product_item_id VARCHAR(32) NOT NULL REFERENCES product_items (id) ON DELETE CASCADE,
    shop_order_id   VARCHAR(32) NOT NULL REFERENCES shop_orders   (id) ON DELETE RESTRICT,
    qty             INT         NOT NULL,
    price_amount_minor BIGINT   NOT NULL DEFAULT 0,
    price_currency   CHAR(3)    NOT NULL DEFAULT 'INR'
);

CREATE INDEX IF NOT EXISTS idx_order_lines_shop_order_id   ON order_lines (shop_order_id);
CREATE INDEX IF NOT EXISTS idx_order_lines_product_item_id ON order_lines (product_item_id);

-- ---------------------------------------------------------------------------
-- 32. ORDER_RETURNS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS order_returns (
    id                       VARCHAR(32) PRIMARY KEY,
    shop_order_id            VARCHAR(32) NOT NULL UNIQUE REFERENCES shop_orders (id) ON DELETE RESTRICT,
    request_date             TIMESTAMPTZ NOT NULL,
    return_reason            TEXT        NOT NULL,
    refund_amount_amount_minor BIGINT     NOT NULL DEFAULT 0,
    refund_amount_currency    CHAR(3)    NOT NULL DEFAULT 'INR',
    is_approved              BOOLEAN     NOT NULL DEFAULT FALSE,
    return_date              TIMESTAMPTZ,
    approval_date            TIMESTAMPTZ,
    admin_comment            TEXT
);

CREATE INDEX IF NOT EXISTS idx_order_returns_shop_order_id ON order_returns (shop_order_id);

-- ---------------------------------------------------------------------------
-- 33. SHOP_OFFERS (many2many: shop_details <-> offers)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shop_offers (
    id         VARCHAR(32) PRIMARY KEY,
    shop_id    VARCHAR(32) NOT NULL REFERENCES shop_details (id) ON DELETE CASCADE,
    offer_id   VARCHAR(32) NOT NULL REFERENCES offers       (id) ON DELETE CASCADE,
    admin_id   VARCHAR(32),
    start_date TIMESTAMPTZ NOT NULL,
    end_date   TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_offers_shop_id  ON shop_offers (shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_offers_offer_id ON shop_offers (offer_id);

-- ---------------------------------------------------------------------------
-- 34. SHOP_DEPARTMENTS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shop_departments (
    id              VARCHAR(32) PRIMARY KEY,
    admin_id        VARCHAR(32) NOT NULL REFERENCES admins        (id) ON DELETE CASCADE,
    shop_id         VARCHAR(32) NOT NULL REFERENCES shop_details  (id) ON DELETE CASCADE,
    department_id   VARCHAR(32) NOT NULL REFERENCES departments   (id) ON DELETE RESTRICT,
    category_id     VARCHAR(32) NOT NULL REFERENCES categories    (id) ON DELETE RESTRICT,
    sub_category_id VARCHAR(32) NOT NULL REFERENCES sub_categories (id) ON DELETE RESTRICT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_shop_departments_shop_dept_cat_sub
    ON shop_departments (shop_id, department_id, category_id, sub_category_id);
CREATE INDEX IF NOT EXISTS idx_shop_departments_admin_id ON shop_departments (admin_id);

-- ---------------------------------------------------------------------------
-- 35. SHOP_TIMES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shop_times (
    id         VARCHAR(32) PRIMARY KEY,
    shop_id    VARCHAR(32) NOT NULL REFERENCES shop_details (id) ON DELETE CASCADE,
    status     VARCHAR(20) NOT NULL
        CHECK (status IN ('open','close')),
    open_time  VARCHAR(50) NOT NULL,
    close_time VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_times_shop_id ON shop_times (shop_id);

-- ---------------------------------------------------------------------------
-- 36. SHOP_SOCIALS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shop_socials (
    id          VARCHAR(32) PRIMARY KEY,
    shop_id     VARCHAR(32) NOT NULL REFERENCES shop_details (id) ON DELETE CASCADE,
    admin_id    VARCHAR(32) NOT NULL REFERENCES admins       (id) ON DELETE CASCADE,
    user_id     VARCHAR(32) NOT NULL REFERENCES users        (id) ON DELETE CASCADE,
    is_follower BOOLEAN     NOT NULL DEFAULT FALSE,
    is_liked    BOOLEAN     NOT NULL DEFAULT FALSE,
    rating      INT         NOT NULL DEFAULT 0
        CHECK (rating BETWEEN 0 AND 5),
    review      TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_socials_shop_id  ON shop_socials (shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_socials_user_id  ON shop_socials (user_id);
CREATE INDEX IF NOT EXISTS idx_shop_socials_admin_id ON shop_socials (admin_id);

-- ---------------------------------------------------------------------------
-- 37. SHOP_VERIFICATIONS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shop_verifications (
    id                   VARCHAR(32) PRIMARY KEY,
    admin_id             VARCHAR(32) NOT NULL REFERENCES admins       (id) ON DELETE RESTRICT,
    shop_id              VARCHAR(32) REFERENCES shop_details          (id) ON DELETE SET NULL,
    shop_name            VARCHAR(255),
    verification_status  BOOLEAN     NOT NULL DEFAULT FALSE,
    remarks              TEXT,
    agent_id             VARCHAR(32),
    created_by           VARCHAR(32),
    updated_by           VARCHAR(32),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_shop_verifications_admin_id ON shop_verifications (admin_id);
CREATE INDEX IF NOT EXISTS idx_shop_verifications_admin_id ON shop_verifications (admin_id);
CREATE INDEX IF NOT EXISTS idx_shop_verifications_shop_id  ON shop_verifications (shop_id);

-- ---------------------------------------------------------------------------
-- 38. SHOP_VERIFICATION_HISTORIES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shop_verification_histories (
    id                  VARCHAR(32) PRIMARY KEY,
    admin_id            VARCHAR(32) NOT NULL REFERENCES admins       (id) ON DELETE RESTRICT,
    shop_id             VARCHAR(32) NOT NULL REFERENCES shop_details (id) ON DELETE CASCADE,
    verification_status VARCHAR(20) NOT NULL
        CHECK (verification_status IN ('verified','unverified','under_review')),
    remarks             VARCHAR(255),
    changed_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_verif_hist_admin_id ON shop_verification_histories (admin_id);
CREATE INDEX IF NOT EXISTS idx_shop_verif_hist_shop_id  ON shop_verification_histories (shop_id);

-- ---------------------------------------------------------------------------
-- 39. ADVERTISEMENTS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS advertisements (
    id               VARCHAR(32) PRIMARY KEY,
    title            VARCHAR(100) NOT NULL,
    content          TEXT         NOT NULL,
    image_url        VARCHAR(255),
    target_url       VARCHAR(255),
    start_date       TIMESTAMPTZ  NOT NULL,
    end_date         TIMESTAMPTZ  NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by_admin VARCHAR(32)  NOT NULL,
    admin_id         VARCHAR(32)  REFERENCES admins (id) ON DELETE SET NULL,
    area_targeted    VARCHAR(255),
    pincode_targeted VARCHAR(20),
    latitude         DECIMAL(10,7),
    longitude        DECIMAL(10,7),
    distance_km      DECIMAL(10,2),
    status           VARCHAR(50)
        CHECK (status IN ('active','inactive','expired') OR status IS NULL),
    priority         VARCHAR(20)
        CHECK (priority IN ('high','medium','low') OR priority IS NULL),
    created_by       VARCHAR(32),
    updated_by       VARCHAR(32)
);

CREATE INDEX IF NOT EXISTS idx_advertisements_admin_id ON advertisements (admin_id);
CREATE INDEX IF NOT EXISTS idx_advertisements_status   ON advertisements (status);

-- ---------------------------------------------------------------------------
-- 40. BANNERS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS banners (
    id            VARCHAR(32) PRIMARY KEY,
    title         VARCHAR(255) NOT NULL,
    description   VARCHAR(500),
    image_url     VARCHAR(500),
    link          VARCHAR(500),
    active        BOOLEAN      NOT NULL DEFAULT TRUE,
    department_id VARCHAR(32)  REFERENCES departments (id) ON DELETE SET NULL,
    category_id   VARCHAR(32)  REFERENCES categories  (id) ON DELETE SET NULL,
    app_type      VARCHAR(50)
        CHECK (app_type IN ('customer','seller') OR app_type IS NULL),
    latitude      DECIMAL(10,7) NOT NULL DEFAULT 0,
    longitude     DECIMAL(10,7) NOT NULL DEFAULT 0,
    pincode       VARCHAR(20),
    radius_km     DECIMAL(10,2) NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_banners_department_id ON banners (department_id);
CREATE INDEX IF NOT EXISTS idx_banners_category_id   ON banners (category_id);

-- ---------------------------------------------------------------------------
-- 41. PROMOTION_CATEGORIES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS promotion_categories (
    id         VARCHAR(32) PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    shop_id    VARCHAR(32)  REFERENCES shop_details (id) ON DELETE SET NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    icon_path  TEXT,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_promotion_categories_shop_id ON promotion_categories (shop_id);

-- ---------------------------------------------------------------------------
-- 42. PROMOTIONS_TYPES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS promotions_types (
    id                      VARCHAR(32) PRIMARY KEY,
    name                    VARCHAR(255),
    is_active               BOOLEAN     NOT NULL DEFAULT TRUE,
    shop_id                 VARCHAR(32) REFERENCES shop_details        (id) ON DELETE SET NULL,
    promotion_category_id   VARCHAR(32) REFERENCES promotion_categories (id) ON DELETE SET NULL,
    promotion_offer_details JSONB       NOT NULL DEFAULT '{}',
    icon_path               TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_promotions_types_shop_id              ON promotions_types (shop_id);
CREATE INDEX IF NOT EXISTS idx_promotions_types_promotion_category_id ON promotions_types (promotion_category_id);

-- ---------------------------------------------------------------------------
-- 43. PROMOTIONS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS promotions (
    id                        VARCHAR(32) PRIMARY KEY,
    promotion_category_id     VARCHAR(32) REFERENCES promotion_categories (id) ON DELETE SET NULL,
    promotion_type_id         VARCHAR(32) REFERENCES promotions_types     (id) ON DELETE SET NULL,
    offer_name                VARCHAR(255),
    description               TEXT,
    discount_rate             DECIMAL(10,2),
    start_date                TIMESTAMPTZ,
    end_date                  TIMESTAMPTZ,
    minimum_purchase_amount   DECIMAL(10,2),
    tier_quantity             INT,
    bogo_get_quantity         INT,
    bogo_buy_quantity         INT,
    bogo_combination_enabled  BOOLEAN,
    gift_description          TEXT,
    shop_id                   VARCHAR(32) REFERENCES shop_details (id) ON DELETE CASCADE,
    is_active                 BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_promotions_shop_id              ON promotions (shop_id);
CREATE INDEX IF NOT EXISTS idx_promotions_promotion_category_id ON promotions (promotion_category_id);

-- ---------------------------------------------------------------------------
-- 44. SUBSCRIPTION_PLANS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS subscription_plans (
    id                          VARCHAR(32) PRIMARY KEY,
    name                        VARCHAR(255) NOT NULL UNIQUE,
    price_monthly_amount_minor   BIGINT       NOT NULL DEFAULT 0,
    price_monthly_currency       CHAR(3)      NOT NULL DEFAULT 'INR',
    duration_days               INT          NOT NULL DEFAULT 30,
    is_active                   BOOLEAN      NOT NULL DEFAULT TRUE
);

-- ---------------------------------------------------------------------------
-- 45. SUBSCRIPTION_ORDERS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS subscription_orders (
    id                  VARCHAR(32) PRIMARY KEY,
    user_id             VARCHAR(32) NOT NULL REFERENCES users              (id) ON DELETE RESTRICT,
    plan_id             VARCHAR(32) NOT NULL REFERENCES subscription_plans (id) ON DELETE RESTRICT,
    price_amount_minor   BIGINT      NOT NULL DEFAULT 0,
    price_currency       CHAR(3)     NOT NULL DEFAULT 'INR',
    razorpay_order_id   VARCHAR(255) NOT NULL UNIQUE,
    razorpay_payment_id VARCHAR(255) UNIQUE,
    status              VARCHAR(20) NOT NULL DEFAULT 'created'
        CHECK (status IN ('created','paid','failed','expired')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at             TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_subscription_orders_user_id  ON subscription_orders (user_id);
CREATE INDEX IF NOT EXISTS idx_subscription_orders_plan_id  ON subscription_orders (plan_id);

-- ---------------------------------------------------------------------------
-- 46. USER_SUBSCRIPTIONS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_subscriptions (
    id                    VARCHAR(32) PRIMARY KEY,
    user_id               VARCHAR(32) NOT NULL REFERENCES users               (id) ON DELETE RESTRICT,
    plan_id               VARCHAR(32) NOT NULL REFERENCES subscription_plans  (id) ON DELETE RESTRICT,
    subscription_order_id VARCHAR(32)          REFERENCES subscription_orders (id) ON DELETE SET NULL,
    is_trial              BOOLEAN      NOT NULL DEFAULT FALSE,
    start_date            TIMESTAMPTZ  NOT NULL,
    end_date              TIMESTAMPTZ  NOT NULL,
    is_active             BOOLEAN      NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user_id ON user_subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_plan_id ON user_subscriptions (plan_id);

-- ---------------------------------------------------------------------------
-- 47. NOTIFICATIONS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notifications (
    id                     VARCHAR(32)  PRIMARY KEY,
    sender_type            VARCHAR(50)  NOT NULL
        CHECK (sender_type IN ('user','admin')),
    receiver_type          VARCHAR(50)  NOT NULL
        CHECK (receiver_type IN ('user','admin')),
    type                   VARCHAR(100) NOT NULL,
    sender_id              VARCHAR(32)  NOT NULL,
    title                  VARCHAR(255) NOT NULL,
    message                TEXT         NOT NULL,
    body                   TEXT         NOT NULL,
    is_read                BOOLEAN      NOT NULL DEFAULT FALSE,
    receiver_id            VARCHAR(32)  NOT NULL,
    category_id            VARCHAR(32)  NOT NULL,
    product_id             VARCHAR(32),
    variation_id           VARCHAR(32),
    shop_id                VARCHAR(32),
    user_id                VARCHAR(32),
    admin_id               VARCHAR(32),
    order_id               VARCHAR(32),
    offer_id               VARCHAR(32),
    notification_meta_data TEXT,
    status                 VARCHAR(50)  NOT NULL
        CHECK (status IN ('pending','sent','read','failed')),
    created_at             VARCHAR(50)  NOT NULL,
    updated_at             VARCHAR(50)  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_notifications_receiver_id ON notifications (receiver_id);
CREATE INDEX IF NOT EXISTS idx_notifications_sender_id   ON notifications (sender_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id     ON notifications (user_id);

-- ---------------------------------------------------------------------------
-- 48. NOTIFICATION_DEVICE_TOKENS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notification_device_tokens (
    id         VARCHAR(32)  PRIMARY KEY,
    owner_id   VARCHAR(100) NOT NULL,
    shop_id    VARCHAR(100) NOT NULL,
    admin_id   VARCHAR(100) NOT NULL,
    owner_type VARCHAR(10)  NOT NULL
        CHECK (owner_type IN ('user','seller')),
    token      VARCHAR(255) NOT NULL UNIQUE,
    platform   VARCHAR(50),
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_notification_device_tokens_owner_id ON notification_device_tokens (owner_id);

-- ---------------------------------------------------------------------------
-- 49. FCM_TOKENS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS fcm_tokens (
    id         VARCHAR(32)  PRIMARY KEY,
    token      VARCHAR(255) NOT NULL UNIQUE,
    device     VARCHAR(255),
    platform   VARCHAR(255),
    owner_id   VARCHAR(255) NOT NULL,
    owner_type VARCHAR(255) NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_fcm_tokens_owner_id ON fcm_tokens (owner_id);

-- ---------------------------------------------------------------------------
-- 50. ALERTS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS alerts (
    id           VARCHAR(32) PRIMARY KEY,
    seller_id    VARCHAR(32) NOT NULL,
    key          VARCHAR(100) NOT NULL,
    title        VARCHAR(255) NOT NULL,
    content      TEXT,
    description  TEXT,
    type         VARCHAR(50)  NOT NULL DEFAULT 'info'
        CHECK (type IN ('info','warning','critical')),
    priority     INT          NOT NULL DEFAULT 0,
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    valid_from   TIMESTAMPTZ,
    valid_until  TIMESTAMPTZ,
    metadata     JSONB        NOT NULL DEFAULT '{}',
    frequency    VARCHAR(50),
    last_shown_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_seller_id  ON alerts (seller_id);
CREATE INDEX IF NOT EXISTS idx_alerts_key        ON alerts (key);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts (created_at);

-- ---------------------------------------------------------------------------
-- 51. ALERT_ACTIONS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS alert_actions (
    id          VARCHAR(32) PRIMARY KEY,
    alert_id    VARCHAR(32) NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    label       VARCHAR(100) NOT NULL,
    action_url  VARCHAR(500),
    action_type VARCHAR(50)  NOT NULL
        CHECK (action_type IN ('navigate','api_call','dismiss')),
    payload     JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alert_actions_alert_id ON alert_actions (alert_id);

-- ---------------------------------------------------------------------------
-- 52. ALERT_TEMPLATES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS alert_templates (
    id               VARCHAR(32) PRIMARY KEY,
    key              VARCHAR(100) NOT NULL UNIQUE,
    title            VARCHAR(255) NOT NULL,
    description      TEXT,
    type             VARCHAR(50)  NOT NULL DEFAULT 'info',
    priority         INT          NOT NULL DEFAULT 0,
    is_active        BOOLEAN      NOT NULL DEFAULT TRUE,
    conditions       JSONB,
    actions          JSONB,
    frequency_config JSONB,
    validity         JSONB,
    enabled          BOOLEAN      NOT NULL DEFAULT TRUE,
    template_type    VARCHAR(50)  NOT NULL DEFAULT 'announcement',
    display_type     VARCHAR(50)  NOT NULL DEFAULT 'bottom_sheet'
        CHECK (display_type IN ('bottom_sheet','banner','modal')),
    content_schema   JSONB,
    flow_key         VARCHAR(100),
    created_by       VARCHAR(32),
    updated_by       VARCHAR(32),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alert_templates_is_active ON alert_templates (is_active);
CREATE INDEX IF NOT EXISTS idx_alert_templates_flow_key  ON alert_templates (flow_key);

-- ---------------------------------------------------------------------------
-- 53. SELLER_ALERT_LOGS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS seller_alert_logs (
    id         VARCHAR(32) PRIMARY KEY,
    seller_id  VARCHAR(32) NOT NULL,
    alert_key  VARCHAR(100) NOT NULL,
    action     VARCHAR(50)  NOT NULL,
    action_url VARCHAR(500),
    metadata   JSONB,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_seller_alert_logs_seller_id  ON seller_alert_logs (seller_id);
CREATE INDEX IF NOT EXISTS idx_seller_alert_logs_created_at ON seller_alert_logs (created_at DESC);

-- ---------------------------------------------------------------------------
-- 54. JOB TABLES
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_categories (
    id          VARCHAR(32) PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    category_id VARCHAR(32),
    parent_id   VARCHAR(32) REFERENCES job_categories (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_job_categories_parent_id ON job_categories (parent_id);

CREATE TABLE IF NOT EXISTS job_sub_categories (
    id          VARCHAR(32) PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    category_id VARCHAR(32) NOT NULL REFERENCES job_categories (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_job_sub_categories_category_id ON job_sub_categories (category_id);

CREATE TABLE IF NOT EXISTS job_locations (
    id   VARCHAR(32) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS job_filters (
    id     VARCHAR(32) PRIMARY KEY,
    name   VARCHAR(100) NOT NULL UNIQUE,
    type   VARCHAR(50)  NOT NULL,
    values TEXT         NOT NULL
);

CREATE TABLE IF NOT EXISTS job_category_filters (
    id          VARCHAR(32) PRIMARY KEY,
    category_id VARCHAR(32) NOT NULL REFERENCES job_categories (id) ON DELETE CASCADE,
    filter_id   VARCHAR(32) NOT NULL REFERENCES job_filters    (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_job_category_filters_category_id ON job_category_filters (category_id);

CREATE TABLE IF NOT EXISTS job_category_locations (
    id          VARCHAR(32) PRIMARY KEY,
    category_id VARCHAR(32) NOT NULL REFERENCES job_categories (id) ON DELETE CASCADE,
    location_id VARCHAR(32) NOT NULL REFERENCES job_locations  (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_job_category_locations_category_id ON job_category_locations (category_id);

CREATE TABLE IF NOT EXISTS jobs (
    id          VARCHAR(32) PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    company     VARCHAR(255) NOT NULL,
    location    VARCHAR(255) NOT NULL,
    salary      BIGINT       NOT NULL,
    category_id VARCHAR(32)  NOT NULL REFERENCES job_categories (id) ON DELETE RESTRICT,
    user_id     VARCHAR(32)  NOT NULL REFERENCES users          (id) ON DELETE RESTRICT,
    posted_date VARCHAR(50)  NOT NULL,
    expiry_date VARCHAR(50)  NOT NULL,
    location_id VARCHAR(32)  NOT NULL REFERENCES job_locations  (id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_jobs_category_id ON jobs (category_id);
CREATE INDEX IF NOT EXISTS idx_jobs_user_id     ON jobs (user_id);
CREATE INDEX IF NOT EXISTS idx_jobs_location_id ON jobs (location_id);

-- ---------------------------------------------------------------------------
-- 55. MOBILE AUTH: MOBILE_USERS, OTP_REQUESTS, LOGIN_AUDIT_LOGS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mobile_users (
    id         VARCHAR(32) PRIMARY KEY,
    phone      VARCHAR(20)  NOT NULL UNIQUE,
    first_name VARCHAR(100),
    last_name  VARCHAR(100),
    email      VARCHAR(255),
    is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_mobile_users_email      ON mobile_users (email);
CREATE INDEX IF NOT EXISTS idx_mobile_users_deleted_at ON mobile_users (deleted_at);

CREATE TABLE IF NOT EXISTS otp_requests (
    id           VARCHAR(32) PRIMARY KEY,
    phone        VARCHAR(20)  NOT NULL,
    otp_hash     VARCHAR(255) NOT NULL,
    expires_at   TIMESTAMPTZ  NOT NULL,
    attempts     INT          NOT NULL DEFAULT 0,
    max_attempts INT          NOT NULL DEFAULT 3,
    status       VARCHAR(50)  NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','verified','expired','blocked')),
    ip_address   VARCHAR(50),
    user_agent   TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_otp_requests_phone       ON otp_requests (phone);
CREATE INDEX IF NOT EXISTS idx_otp_requests_expires_at  ON otp_requests (expires_at);
CREATE INDEX IF NOT EXISTS idx_otp_requests_status      ON otp_requests (status);
CREATE INDEX IF NOT EXISTS idx_otp_requests_created_at  ON otp_requests (created_at DESC);

CREATE TABLE IF NOT EXISTS login_audit_logs (
    id         VARCHAR(32) PRIMARY KEY,
    phone      VARCHAR(20)  NOT NULL,
    event      VARCHAR(50),
    ip_address VARCHAR(50),
    user_agent TEXT,
    details    JSONB,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_login_audit_logs_phone      ON login_audit_logs (phone);
CREATE INDEX IF NOT EXISTS idx_login_audit_logs_event      ON login_audit_logs (event);
CREATE INDEX IF NOT EXISTS idx_login_audit_logs_created_at ON login_audit_logs (created_at DESC);

-- ---------------------------------------------------------------------------
-- 56. AUTH SESSION TABLES: ADMIN_REFRESH_SESSIONS, USER_REFRESH_SESSIONS,
--     OTP_SESSIONS, OTP_SESSION_EMAILS
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS admin_refresh_sessions (
    token_id      VARCHAR(255) PRIMARY KEY,
    user_id       VARCHAR(32),
    admin_id      VARCHAR(32),
    user_type     VARCHAR(20)
        CHECK (user_type IN ('user','admin') OR user_type IS NULL),
    refresh_token TEXT        NOT NULL,
    expire_at     TIMESTAMPTZ NOT NULL,
    is_blocked    BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_admin_refresh_sessions_user_id  ON admin_refresh_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_admin_refresh_sessions_admin_id ON admin_refresh_sessions (admin_id);

CREATE TABLE IF NOT EXISTS user_refresh_sessions (
    token_id      VARCHAR(255) PRIMARY KEY,
    user_id       VARCHAR(32),
    admin_id      VARCHAR(32),
    user_type     VARCHAR(20)
        CHECK (user_type IN ('user','admin') OR user_type IS NULL),
    refresh_token TEXT        NOT NULL,
    expire_at     TIMESTAMPTZ NOT NULL,
    is_blocked    BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_user_refresh_sessions_user_id  ON user_refresh_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_refresh_sessions_admin_id ON user_refresh_sessions (admin_id);

CREATE TABLE IF NOT EXISTS otp_sessions (
    id         VARCHAR(32) PRIMARY KEY,
    otp_id     VARCHAR(255) NOT NULL UNIQUE,
    user_id    VARCHAR(32),
    admin_id   VARCHAR(32),
    user_type  VARCHAR(20)
        CHECK (user_type IN ('user','admin') OR user_type IS NULL),
    phone      VARCHAR(50),
    expire_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS otp_sessions_email (
    id         VARCHAR(32) PRIMARY KEY,
    otp_id     VARCHAR(255) NOT NULL UNIQUE,
    user_id    VARCHAR(32)  NOT NULL,
    admin_id   VARCHAR(32)  NOT NULL,
    user_type  VARCHAR(20)  NOT NULL
        CHECK (user_type IN ('user','admin')),
    email      VARCHAR(255) NOT NULL,
    expire_at  TIMESTAMPTZ  NOT NULL
);

-- ---------------------------------------------------------------------------
-- 57. USER_CONSENTS (DPDP)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_consents (
    id            VARCHAR(32) PRIMARY KEY,
    user_id       VARCHAR(32) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    consent_type  VARCHAR(20) NOT NULL
        CHECK (consent_type IN ('terms','privacy','marketing')),
    terms_version VARCHAR(50) NOT NULL,
    accepted      BOOLEAN     NOT NULL,
    accepted_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address    VARCHAR(50),
    user_agent    TEXT
);

CREATE INDEX IF NOT EXISTS idx_user_consents_user_id ON user_consents (user_id);

-- ---------------------------------------------------------------------------
-- 58. AUDIT_LOGS (DPDP)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_logs (
    id          VARCHAR(32) PRIMARY KEY,
    actor_id    VARCHAR(32),
    actor_type  VARCHAR(20)
        CHECK (actor_type IN ('user','admin') OR actor_type IS NULL),
    action      VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id   VARCHAR(32),
    ip_address  VARCHAR(50),
    metadata    JSONB,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id    ON audit_logs (actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action      ON audit_logs (action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_type ON audit_logs (entity_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_id   ON audit_logs (entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at  ON audit_logs (created_at DESC);

-- Recreate trigger functions
CREATE OR REPLACE FUNCTION update_cart_total_price() 
RETURNS TRIGGER AS $$ 
BEGIN 
IF (TG_OP = 'DELETE') THEN
	UPDATE carts c
		SET total_price_amount_minor = (
			SELECT COALESCE ( SUM ( CASE WHEN pi.discount_price > 0 THEN pi.discount_price * ci.qty ELSE pi.price * ci.qty END), 0)::bigint
			FROM cart_items ci INNER JOIN product_items pi ON ci.product_item_id = pi.id
			WHERE ci.cart_id = OLD.cart_id
		), applied_coupon_id = 0, discount_amount_amount_minor = 0
	WHERE c.id = OLD.cart_id;
	RETURN NEW;
ELSE
	UPDATE carts c
		SET total_price_amount_minor = (
			SELECT SUM (CASE WHEN pi.discount_price > 0 THEN pi.discount_price * ci.qty ELSE pi.price * ci.qty END)
			FROM cart_items ci INNER JOIN product_items pi ON ci.product_item_id = pi.id
			WHERE ci.cart_id = NEW.cart_id
		), applied_coupon_id = 0, discount_amount_amount_minor = 0
		WHERE c.id = NEW.cart_id;

END IF; 
RETURN NEW; 
END; 
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_product_quantity() 
RETURNS TRIGGER AS $$ 
BEGIN 
	IF (TG_OP = 'INSERT') THEN 
		UPDATE product_items pi 
		SET qty_in_stock = pi.qty_in_stock - NEW.qty 
		WHERE pi.id = NEW.product_item_id; 

	END IF; 
	RETURN NEW; 
END; 
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_product_quantity_on_return()
RETURNS TRIGGER AS $$
BEGIN
  IF (TG_OP = 'UPDATE' AND NEW.status = 'order returned') THEN 
	UPDATE product_items pi
	SET qty_in_stock = pi.qty_in_stock + ol.qty
	FROM order_lines ol
	WHERE pi.id = ol.product_item_id
	AND ol.shop_order_id = NEW.id;

	RETURN NEW;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate triggers
CREATE OR REPLACE TRIGGER update_cart_total_price
AFTER INSERT OR UPDATE OR DELETE ON cart_items
FOR EACH ROW EXECUTE FUNCTION update_cart_total_price();

CREATE OR REPLACE TRIGGER update_product_quantity 
AFTER INSERT ON order_lines 
FOR EACH ROW EXECUTE FUNCTION update_product_quantity();

CREATE OR REPLACE TRIGGER update_product_qty_on_order_return
AFTER UPDATE ON shop_orders
FOR EACH ROW
EXECUTE FUNCTION update_product_quantity_on_return();
