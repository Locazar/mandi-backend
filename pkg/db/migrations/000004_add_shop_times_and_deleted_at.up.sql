CREATE TABLE IF NOT EXISTS shop_times (
    id         VARCHAR(32) PRIMARY KEY,
    shop_id    VARCHAR(32) NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'close',
    open_time  VARCHAR(20) NOT NULL DEFAULT '',
    close_time VARCHAR(20) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_times_shop_id ON shop_times (shop_id);

ALTER TABLE shop_details ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_shop_details_deleted_at ON shop_details (deleted_at);
