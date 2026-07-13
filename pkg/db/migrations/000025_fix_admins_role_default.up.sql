-- admins.role defaulted to 'super_admin' (migration 000007), and SaveAdmin's
-- raw INSERT never included the role column at all, so every seller/customer
-- self-signup silently got 'super_admin' from that column default. role is a
-- platform-user-only concept — assigned by a super_admin via CreateAdmin —
-- and must never be set on a self-registered seller/customer account, so the
-- default now matches that: '' (blank), not 'seller' and not 'super_admin'.
-- Fixed in code (pkg/repository/admin.go, pkg/usecase/admin.go,
-- pkg/usecase/auth.go) to always pass role explicitly (blank for
-- sellers/customers); this migration removes the dangerous default so any
-- future insert path that forgets to set role fails safe (no platform
-- access) instead of fails open (super_admin).
ALTER TABLE admins ALTER COLUMN role SET DEFAULT '';

-- One-time data repair: every admin row auto-created by seller/customer
-- self-signup has a username matching GenerateRandomUserName("seller") —
-- literally "seller" + 4 digits (pkg/utils/helper_functions.go) — and picked
-- up role='super_admin' from the old column default purely as a side effect
-- of the missing INSERT column, not because anyone intended these accounts
-- to be platform admins. Null out role only for rows matching that exact
-- auto-generated pattern; real platform users (created via CreateAdmin, with
-- deliberately chosen usernames like "superadmin"/"marketing"/"support")
-- never match this pattern and are left untouched.
UPDATE admins
SET role = ''
WHERE role = 'super_admin'
  AND user_name ~ '^seller[0-9]{4}$';
