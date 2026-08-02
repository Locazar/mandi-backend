-- QR redirect codes. An admin creates a short code that resolves on our server
-- and 302-redirects the scanner to a target URL (Play Store, YouTube, …).
--
-- "Dynamic" by design: target_url is an editable column, so a printed/shared QR
-- keeps working forever and the destination can be repointed without reprinting.
-- Purely additive — no existing table is touched.

CREATE TABLE IF NOT EXISTS qr_codes (
    id              VARCHAR(32)  PRIMARY KEY,
    -- Public lookup key embedded in the short link (/r/<short_code>). Short and
    -- URL-safe so the encoded QR stays low-density and easy to scan.
    short_code      VARCHAR(16)  NOT NULL,
    title           VARCHAR(150) NOT NULL DEFAULT '',
    -- Redirect destination. Validated to http/https at the API layer.
    target_url      TEXT         NOT NULL,
    -- Disabled links resolve to HTTP 410 instead of redirecting.
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    -- Optional hard expiry; NULL means never expires.
    expires_at      TIMESTAMPTZ  NULL,
    scan_count      BIGINT       NOT NULL DEFAULT 0,
    last_scanned_at TIMESTAMPTZ  NULL,
    -- Admin id that created it (nullable — informational only).
    created_by      VARCHAR(32)  NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    -- Soft delete: preserves short_code (never reused) and scan history.
    deleted_at      TIMESTAMPTZ  NULL
);

-- short_code must be unique among live rows; a soft-deleted code is retired, not
-- reused, so the partial predicate keeps the invariant without blocking deletes.
CREATE UNIQUE INDEX IF NOT EXISTS uidx_qr_codes_short_code
    ON qr_codes (short_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_qr_codes_deleted_at ON qr_codes (deleted_at);
CREATE INDEX IF NOT EXISTS idx_qr_codes_created_at ON qr_codes (created_at DESC);

-- One row per scan, for analytics/reporting. Best-effort: a failed insert here
-- never blocks the redirect.
CREATE TABLE IF NOT EXISTS qr_scan_events (
    id         VARCHAR(32) PRIMARY KEY,
    qr_code_id VARCHAR(32) NOT NULL REFERENCES qr_codes (id) ON DELETE CASCADE,
    scanned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address VARCHAR(64),
    user_agent TEXT,
    referer    TEXT
);

CREATE INDEX IF NOT EXISTS idx_qr_scan_events_qr_time
    ON qr_scan_events (qr_code_id, scanned_at DESC);
