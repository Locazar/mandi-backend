-- 000002_social_module.up.sql

-- Create ENUM types
CREATE TYPE review_type AS ENUM (
    'SHOP',
    'PRODUCT'
);

CREATE TYPE reaction_type AS ENUM (
    'LIKE',
    'HELPFUL',
    'REPORT_ABUSE'
);

CREATE TYPE feedback_type AS ENUM (
    'SHOP',
    'PRODUCT',
    'APP'
);

CREATE TYPE review_status AS ENUM (
    'active',
    'deleted',
    'moderated'
);

CREATE TYPE feedback_status AS ENUM (
    'new',
    'resolved',
    'archived'
);

-- shop_reviews Table
CREATE TABLE IF NOT EXISTS shop_reviews (
    id              SERIAL PRIMARY KEY,
    shop_id         VARCHAR(32)  NOT NULL,
    user_id         VARCHAR(32)  NOT NULL,
    rating          DECIMAL(2,1) NOT NULL CHECK (rating >= 1.0 AND rating <= 5.0),
    review_text     TEXT         CHECK (LENGTH(review_text) <= 2000),
    status          review_status NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT fk_shop_reviews_shop_id FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE,
    CONSTRAINT fk_shop_reviews_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uix_shop_reviews_shop_id_user_id ON shop_reviews (shop_id, user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_shop_reviews_shop_id    ON shop_reviews (shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_reviews_user_id    ON shop_reviews (user_id);
CREATE INDEX IF NOT EXISTS idx_shop_reviews_deleted_at ON shop_reviews (deleted_at);

-- product_reviews Table
CREATE TABLE IF NOT EXISTS product_reviews (
    id              SERIAL PRIMARY KEY,
    product_item_id VARCHAR(32) NOT NULL,
    shop_id         VARCHAR(32) NOT NULL,
    user_id         VARCHAR(32) NOT NULL,
    rating          DECIMAL(2,1) NOT NULL CHECK (rating >= 1.0 AND rating <= 5.0),
    review_text     TEXT         CHECK (LENGTH(review_text) <= 2000),
    status          review_status NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT fk_product_reviews_product_item_id FOREIGN KEY (product_item_id) REFERENCES product_items(id) ON DELETE CASCADE,
    CONSTRAINT fk_product_reviews_shop_id         FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE,
    CONSTRAINT fk_product_reviews_user_id         FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uix_product_reviews_product_item_id_user_id ON product_reviews (product_item_id, user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_product_reviews_product_item_id ON product_reviews (product_item_id);
CREATE INDEX IF NOT EXISTS idx_product_reviews_shop_id         ON product_reviews (shop_id);
CREATE INDEX IF NOT EXISTS idx_product_reviews_user_id         ON product_reviews (user_id);
CREATE INDEX IF NOT EXISTS idx_product_reviews_deleted_at      ON product_reviews (deleted_at);

-- review_images Table
CREATE TABLE IF NOT EXISTS review_images (
    id            SERIAL PRIMARY KEY,
    review_type   review_type NOT NULL,
    review_id     INT         NOT NULL,
    image_url     TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- No direct foreign key due to 'review_type' abstraction, handled in application logic
    -- Potentially add CHECK constraints if needed for review_type and review_id pairing
);

CREATE INDEX IF NOT EXISTS idx_review_images_review_type_id ON review_images (review_type, review_id);

-- review_comments Table
CREATE TABLE IF NOT EXISTS review_comments (
    id                  SERIAL PRIMARY KEY,
    review_type         review_type   NOT NULL,
    review_id           INT           NOT NULL,
    parent_comment_id   INT,
    user_id             VARCHAR(32)   NOT NULL,
    comment             TEXT          NOT NULL CHECK (LENGTH(comment) <= 1000),
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT fk_review_comments_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_review_comments_parent_id FOREIGN KEY (parent_comment_id) REFERENCES review_comments(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_review_comments_review_type_id ON review_comments (review_type, review_id);
CREATE INDEX IF NOT EXISTS idx_review_comments_user_id        ON review_comments (user_id);
CREATE INDEX IF NOT EXISTS idx_review_comments_parent_id      ON review_comments (parent_comment_id);
CREATE INDEX IF NOT EXISTS idx_review_comments_deleted_at     ON review_comments (deleted_at);

-- review_reactions Table
CREATE TABLE IF NOT EXISTS review_reactions (
    id              SERIAL PRIMARY KEY,
    review_type     review_type   NOT NULL,
    review_id       INT           NOT NULL,
    user_id         VARCHAR(32)   NOT NULL,
    reaction_type   reaction_type NOT NULL,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_review_reactions_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uix_review_reactions_review_user_type UNIQUE (review_type, review_id, user_id, reaction_type) -- A user can only have one of each reaction type per review
);

CREATE INDEX IF NOT EXISTS idx_review_reactions_review_type_id ON review_reactions (review_type, review_id);
CREATE INDEX IF NOT EXISTS idx_review_reactions_user_id        ON review_reactions (user_id);

-- feedbacks Table
CREATE TABLE IF NOT EXISTS feedbacks (
    id              SERIAL PRIMARY KEY,
    feedback_type   feedback_type NOT NULL,
    shop_id         VARCHAR(32),
    product_item_id VARCHAR(32),
    user_id         VARCHAR(32),
    rating          DECIMAL(2,1)  CHECK (rating >= 1.0 AND rating <= 5.0),
    feedback_text   TEXT          NOT NULL,
    status          feedback_status NOT NULL DEFAULT 'new',
    anonymous       BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_feedbacks_shop_id FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE SET NULL,
    CONSTRAINT fk_feedbacks_product_item_id FOREIGN KEY (product_item_id) REFERENCES product_items(id) ON DELETE SET NULL,
    CONSTRAINT fk_feedbacks_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_feedbacks_feedback_type ON feedbacks (feedback_type);
CREATE INDEX IF NOT EXISTS idx_feedbacks_shop_id       ON feedbacks (shop_id);
CREATE INDEX IF NOT EXISTS idx_feedbacks_product_item_id ON feedbacks (product_item_id);
CREATE INDEX IF NOT EXISTS idx_feedbacks_user_id       ON feedbacks (user_id);

-- shop_rating_summary Table
CREATE TABLE IF NOT EXISTS shop_rating_summary (
    id              SERIAL PRIMARY KEY,
    shop_id         VARCHAR(32)  NOT NULL UNIQUE,
    average_rating  DECIMAL(2,1) NOT NULL DEFAULT 0.0,
    total_reviews   INT          NOT NULL DEFAULT 0,
    rating_1_count  INT          NOT NULL DEFAULT 0,
    rating_2_count  INT          NOT NULL DEFAULT 0,
    rating_3_count  INT          NOT NULL DEFAULT 0,
    rating_4_count  INT          NOT NULL DEFAULT 0,
    rating_5_count  INT          NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_shop_rating_summary_shop_id FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_shop_rating_summary_shop_id ON shop_rating_summary (shop_id);

-- product_rating_summary Table
CREATE TABLE IF NOT EXISTS product_rating_summary (
    id                SERIAL PRIMARY KEY,
    product_item_id   VARCHAR(32)  NOT NULL UNIQUE,
    average_rating    DECIMAL(2,1) NOT NULL DEFAULT 0.0,
    total_reviews     INT          NOT NULL DEFAULT 0,
    rating_1_count    INT          NOT NULL DEFAULT 0,
    rating_2_count    INT          NOT NULL DEFAULT 0,
    rating_3_count    INT          NOT NULL DEFAULT 0,
    rating_4_count    INT          NOT NULL DEFAULT 0,
    rating_5_count    INT          NOT NULL DEFAULT 0,
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_product_rating_summary_product_item_id FOREIGN KEY (product_item_id) REFERENCES product_items(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_product_rating_summary_product_item_id ON product_rating_summary (product_item_id);

