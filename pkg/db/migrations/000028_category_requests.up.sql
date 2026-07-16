-- Lets a seller request a department/category that isn't in the existing
-- list (surfaced from the seller-app's "More" section), for a super_admin
-- or catalog manager to review from admin-portal and act on (create the
-- real department/category, or reject with a note).
CREATE TABLE IF NOT EXISTS category_requests (
    id              VARCHAR(32)  PRIMARY KEY,
    admin_id        VARCHAR(32)  NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    shop_id         VARCHAR(32)  REFERENCES shop_details(id) ON DELETE SET NULL,
    department_name VARCHAR(100) NOT NULL,
    category_name   VARCHAR(100),
    note            TEXT,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending'
                                  CHECK (status IN ('pending', 'approved', 'rejected')),
    admin_response  TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_category_requests_admin_id ON category_requests (admin_id);
CREATE INDEX IF NOT EXISTS idx_category_requests_status ON category_requests (status);
