# Baseline DB Seed & Migration Verification

**Date:** 2026-06-04  
**Scope:** mandi-backend only  
**Approach:** Migration-first (Approach A) — drop legacy DB, let migration system rebuild schema, apply reference-data seed

---

## Problem Statement

The current `mandi` database was created with the old legacy schema (bigint PKs, columns like `users.age`, `users.verified`, `admins.aadhar`). The mandi-backend codebase has since migrated to string PKs with typed prefixes (`dept_`, `cat_`, `scat_`, etc.). The live DB holds a lot of user/shop/order data that is no longer needed in development. The goal is to:

1. Replace the legacy DB with a clean, migration-managed schema
2. Populate it with reference data only (departments, categories, sub-categories, variations, sub-type attributes, brands) extracted and remapped from the latest backup
3. Verify the backend starts and serves correctly against the clean DB

---

## Architecture

### Restore Flow

```
1. Start Docker Desktop + docker-compose up -d
2. Drop & recreate mandi database
3. make run  →  RunMigrations (000001–000007) + savePaymentMethods + SeedCountries
4. Ctrl+C
5. psql mandi < baseline_seed.sql
6. make run  →  server on :3000
7. Smoke-test key endpoints
```

### What baseline_seed.sql Contains

Only reference/lookup data — no users, shops, orders, or any PII:

| Table | Source | Rows (approx) |
|---|---|---|
| `departments` | backup | 16 |
| `categories` | backup | ~170 |
| `sub_categories` | backup | ~700 |
| `variations` | backup | varies |
| `variation_options` | backup | varies |
| `sub_type_attributes` | backup | ~2,570 |
| `sub_type_attribute_options` | backup | ~16,194 |
| `brands` | backup | few |

Countries, payment methods, and order statuses are seeded automatically by `make run` (via `SeedCountries`, `savePaymentMethods` in `pkg/db/connection.go`).

---

## Migration System Analysis

### Fresh DB path (what we use)

`shouldUseLegacyBootstrap` returns `false` on an empty DB → runs `pkg/db/migrations/` sequence:

| Migration | What it does | Behaviour on fresh DB |
|---|---|---|
| 000001 | Full schema (all tables, string PKs, FK constraints, indexes) | Creates everything |
| 000002 | `users.trial_used` | No-op (baseline already has it) |
| 000003 | `shop_socials` table | No-op (baseline already has it) |
| 000004 | `shop_times` table + `shop_details.deleted_at` | No-op (baseline already has it) |
| 000005 | `admins.aadhaar_last4`, `admins.aadhaar_verified` | No-op (baseline already has them) |
| 000006 | Align `users` (`email_verified`, `phone_verified`, `date_of_birth`, `deleted_at`) | No-op (baseline already has them) |
| 000007 | `admins.role` | **Actually runs** — baseline omits `role` intentionally |

**Result:** No migration code changes needed. The sequence is correct and idempotent. Migrations 002–006 exist for upgrading legacy DBs; they are safe no-ops on fresh DBs due to `IF NOT EXISTS`.

### Legacy DB path (preserved but not used for fresh installs)

`shouldUseLegacyBootstrap` detects old bigint PKs or presence of `refresh_sessions` → runs `pkg/db/legacy_migrations/000001_legacy_backup_compat.up.sql`, which converts bigint IDs to varchar strings. This path is preserved for anyone who needs to upgrade an existing legacy install.

---

## Seed Generation Script

### Script: `scripts/generate_baseline_seed.py`

**Input:** `mandi_backup_20260604_074125.sql`  
**Output:** `baseline_seed.sql`

**Algorithm:**

1. Parse COPY blocks for target tables from the backup file
2. Build ID mapping tables: old numeric ID → new prefixed string ID
   - `departments`: `1` → `dept_0001`, `2` → `dept_0002`, …
   - `categories`: `2` → `cat_0002`, `3` → `cat_0003`, …
   - `sub_categories`: `16` → `scat_0016`, …
   - `variations`: `var_XXXXX`
   - `variation_options`: `varo_XXXXX`
   - `sub_type_attributes`: `sta_XXXXX`
   - `sub_type_attribute_options`: `stao_XXXXX`
   - `brands`: `brnd_XXXX`
3. For each table, remap all FK columns using the ID maps
4. Drop incompatible columns:
   - `departments`: drop `slug`
   - `categories`: drop `slug` (not present in backup) and `category_id`
   - `brands`: drop `slug`
5. Add missing columns as NULL:
   - `sub_categories.icon` (not in backup, nullable in new schema)
6. Emit `INSERT INTO … ON CONFLICT DO NOTHING` statements

**ID format rationale:** `prefix_NNNN` (zero-padded 4–5 digits) is human-readable, traceable to backup row, and won't collide with app-generated IDs (which use cuid2 random suffixes).

---

## Schema Mismatch Findings

Columns in the legacy backup that are **not** in the current Go structs or migration schema:

| Table | Legacy column | Status |
|---|---|---|
| `admins` | `aadhar`, `agree_to_terms`, `department_type`, `shop_name`, `gstin`, `shop_id`, `verified` | Dropped in new schema — handled by legacy migration |
| `users` | `age`, `verified` | Dropped by migration 006 |
| `products` | `stock`, `price`, `discount_price`, `brand_id`, `location_id` | Legacy only — not in new Product struct |
| `departments` | `slug` | Not in Department struct — omit from seed |
| `categories` | `category_id` (self-ref) | Not in Category struct — omit from seed |
| `brands` | `slug` | Not in Brand struct — omit from seed |
| `payment_methods` | `maximum_amount` (single column) | Replaced by `maximum_amount_amount_minor` + `maximum_amount_currency` in new schema |
| `coupons` | `minimum_cart_price` (single column) | Replaced by `minimum_cart_price_amount_minor` + `minimum_cart_price_currency` |

None of these require code changes — the migration system already handles all transitions. The seed script simply omits them.

---

## Verification Plan

After `make run` with the seeded DB:

1. `GET /api/health` → `{"status":"ok"}`
2. `GET /api/departments` or equivalent → returns 16 departments
3. `GET /api/categories` → returns categories with correct department FK
4. `GET /api/sub-categories` → returns sub-categories with correct category FK
5. Admin login endpoint → returns 200 (no crash on empty users table)
6. Check server logs for migration errors on startup

---

## Files Changed

| Path | Action |
|---|---|
| `scripts/generate_baseline_seed.py` | New — generates baseline_seed.sql |
| `baseline_seed.sql` | New — reference data for clean installs |
| `mandi_backup_20260604_074125.sql` | Keep as-is (source data) |
| `mandi_backup.sql`, `backup.sql` | Can be deleted after verification |
