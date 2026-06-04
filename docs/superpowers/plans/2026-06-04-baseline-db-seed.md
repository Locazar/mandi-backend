# Baseline DB Seed Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the legacy mandi database with a clean migration-managed schema populated only with reference data (departments, categories, sub-categories, variations, sub-type attributes, brands).

**Architecture:** A Python script reads the legacy backup, remaps numeric IDs to typed-prefix string IDs, and writes `baseline_seed.sql`. The DB is dropped and recreated, `make run` applies all 7 migrations and auto-seeds countries/payment methods, then `baseline_seed.sql` is applied via psql through Docker.

**Tech Stack:** Python 3 (stdlib only), PostgreSQL 15, golang-migrate, Go 1.21+, Docker Compose

---

## File Map

| Path | Action | Purpose |
|---|---|---|
| `scripts/generate_baseline_seed.py` | Create | Parses backup, remaps IDs, writes INSERT statements |
| `scripts/test_generate_seed.py` | Create | Unit tests for parsing and ID-mapping logic |
| `baseline_seed.sql` | Generated | Reference data INSERT statements (output of script) |
| `pkg/db/migrations/` | Read-only verify | Confirm 000001–000007 are correct — no changes expected |

---

## Task 1: Verify migration sequence is correct

**Files:** Read-only — `pkg/db/migrations/000001_baseline.up.sql` through `000007_add_role_to_admins.up.sql`

- [ ] **Step 1: Confirm each migration is idempotent on a fresh DB**

Read each `.up.sql` file and verify every `ALTER TABLE` uses `IF NOT EXISTS`:

```bash
cd mandi-backend
grep -rn "ALTER TABLE.*ADD COLUMN" pkg/db/migrations/ | grep -v "IF NOT EXISTS"
```

Expected output: **empty** (all `ADD COLUMN` statements use `IF NOT EXISTS`).

- [ ] **Step 2: Confirm baseline includes all struct fields except `admins.role`**

```bash
grep -n "role" pkg/db/migrations/000001_baseline.up.sql
grep -n "role" pkg/db/migrations/000007_add_role_to_admins.up.sql
```

Expected:
- First command: no output (role not in baseline)
- Second command: `ALTER TABLE admins ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'super_admin';`

This confirms the migration chain is complete: baseline creates schema → 007 adds `role`.

- [ ] **Step 3: Confirm no migration code changes are needed**

```bash
grep -rn "ADD COLUMN IF NOT EXISTS" pkg/db/migrations/*.up.sql | wc -l
```

Expected: `8` (one per column added across migrations 002–007). If output matches, migrations are clean. **No edits needed.**

- [ ] **Step 4: Commit verification note**

```bash
git add docs/superpowers/specs/2026-06-04-baseline-db-seed-design.md \
        docs/superpowers/plans/2026-06-04-baseline-db-seed.md
git commit -m "docs: add baseline seed spec and implementation plan"
```

---

## Task 2: Write the seed generation script with tests

**Files:**
- Create: `scripts/test_generate_seed.py`
- Create: `scripts/generate_baseline_seed.py`

- [ ] **Step 1: Write the failing tests first**

Create `scripts/test_generate_seed.py`:

```python
"""Unit tests for generate_baseline_seed helpers."""
import sys, os
sys.path.insert(0, os.path.dirname(__file__))

import generate_baseline_seed as g


def test_make_id_pads_to_5_digits():
    assert g.make_id("dept", "1") == "dept_00001"
    assert g.make_id("cat", "42") == "cat_00042"
    assert g.make_id("sta", "12345") == "sta_12345"


def test_sql_val_null():
    assert g.sql_val("\\N") == "NULL"


def test_sql_val_booleans():
    assert g.sql_val("t") == "TRUE"
    assert g.sql_val("f") == "FALSE"


def test_sql_val_string_escapes_quotes():
    assert g.sql_val("O'Brien") == "'O''Brien'"


def test_sql_val_plain_string():
    assert g.sql_val("Hardware") == "'Hardware'"


def test_sql_val_integer_string():
    # integers stay as strings; Postgres coerces them in INSERT
    assert g.sql_val("1") == "'1'"


def test_parse_copy_block_basic():
    sample = (
        "COPY public.departments (id, name, sort_order) FROM stdin;\n"
        "1\tHardware\t1\n"
        "2\tMaterial\t2\n"
        "\\.\n"
    )
    blocks = g.parse_copy_blocks_from_string(sample)
    assert "departments" in blocks
    cols, rows = blocks["departments"]
    assert cols == ["id", "name", "sort_order"]
    assert len(rows) == 2
    assert rows[0] == ["1", "Hardware", "1"]


def test_parse_copy_block_empty():
    sample = (
        "COPY public.brands (id, name) FROM stdin;\n"
        "\\.\n"
    )
    blocks = g.parse_copy_blocks_from_string(sample)
    cols, rows = blocks["brands"]
    assert rows == []


def test_build_id_map():
    rows = [["1", "Hardware"], ["2", "Material"]]
    m = g.build_id_map(rows, 0, "dept")
    assert m == {"1": "dept_00001", "2": "dept_00002"}


if __name__ == "__main__":
    import unittest
    # run with: python scripts/test_generate_seed.py
    passed = failed = 0
    tests = [v for k, v in globals().items() if k.startswith("test_")]
    for t in tests:
        try:
            t()
            print(f"  PASS  {t.__name__}")
            passed += 1
        except Exception as e:
            print(f"  FAIL  {t.__name__}: {e}")
            failed += 1
    print(f"\n{passed} passed, {failed} failed")
    sys.exit(0 if failed == 0 else 1)
```

- [ ] **Step 2: Run tests — expect failure (module not found)**

```bash
cd mandi-backend
python3 scripts/test_generate_seed.py
```

Expected output:
```
ModuleNotFoundError: No module named 'generate_baseline_seed'
```

- [ ] **Step 3: Write the full implementation**

Create `scripts/generate_baseline_seed.py`:

```python
#!/usr/bin/env python3
"""
Generate baseline_seed.sql from the legacy mandi backup.

Usage:
    python3 scripts/generate_baseline_seed.py

Reads:  mandi_backup_20260604_074125.sql  (in repo root)
Writes: baseline_seed.sql                 (in repo root)
"""
from pathlib import Path

SCRIPT_DIR = Path(__file__).parent
REPO_ROOT = SCRIPT_DIR.parent
BACKUP = REPO_ROOT / "mandi_backup_20260604_074125.sql"
OUTPUT = REPO_ROOT / "baseline_seed.sql"


def make_id(prefix: str, old_id: str) -> str:
    """Map a legacy numeric ID to a new typed-prefix string ID."""
    return f"{prefix}_{int(old_id):05d}"


def sql_val(v: str) -> str:
    """Convert a pg_dump tab-separated field to a SQL literal."""
    if v == "\\N":
        return "NULL"
    if v == "t":
        return "TRUE"
    if v == "f":
        return "FALSE"
    return "'" + v.replace("'", "''") + "'"


def parse_copy_blocks_from_string(text: str) -> dict:
    """Parse COPY blocks from a string (used by tests)."""
    blocks = {}
    current_table = None
    current_cols = None
    current_rows = None

    for raw_line in text.splitlines():
        line = raw_line
        if line.startswith("COPY public.") and "FROM stdin;" in line:
            table = line[len("COPY public."):line.index(" (")]
            cols_str = line[line.index("(") + 1:line.index(")")]
            current_table = table
            current_cols = [c.strip() for c in cols_str.split(",")]
            current_rows = []
        elif line == "\\." and current_table is not None:
            blocks[current_table] = (current_cols, current_rows)
            current_table = None
        elif current_table is not None and line:
            current_rows.append(line.split("\t"))

    return blocks


def parse_copy_blocks(path: Path) -> dict:
    """Parse all COPY blocks from a pg_dump file."""
    blocks = {}
    current_table = None
    current_cols = None
    current_rows = None

    with open(path, encoding="utf-8") as f:
        for raw_line in f:
            line = raw_line.rstrip("\n")
            if line.startswith("COPY public.") and "FROM stdin;" in line:
                table = line[len("COPY public."):line.index(" (")]
                cols_str = line[line.index("(") + 1:line.index(")")]
                current_table = table
                current_cols = [c.strip() for c in cols_str.split(",")]
                current_rows = []
            elif line == "\\." and current_table is not None:
                blocks[current_table] = (current_cols, current_rows)
                current_table = None
            elif current_table is not None and line:
                current_rows.append(line.split("\t"))

    return blocks


def build_id_map(rows: list, id_col_index: int, prefix: str) -> dict:
    """Return {old_id_str: new_prefixed_id} for all non-null IDs."""
    return {
        row[id_col_index]: make_id(prefix, row[id_col_index])
        for row in rows
        if row[id_col_index] != "\\N"
    }


def write_inserts(out, table: str, columns: list, rows: list) -> None:
    """Write INSERT ... ON CONFLICT DO NOTHING for a batch of rows."""
    if not rows:
        out.write(f"-- {table}: no data in backup\n\n")
        return
    col_str = ", ".join(columns)
    out.write(f"INSERT INTO {table} ({col_str}) VALUES\n")
    parts = [
        "  (" + ", ".join(sql_val(v) for v in row) + ")"
        for row in rows
    ]
    out.write(",\n".join(parts))
    out.write("\nON CONFLICT DO NOTHING;\n\n")


def main() -> None:
    print(f"Reading {BACKUP} ...")
    blocks = parse_copy_blocks(BACKUP)

    # --- Extract raw data ---
    dept_cols,  dept_rows  = blocks.get("departments",               ([], []))
    cat_cols,   cat_rows   = blocks.get("categories",                ([], []))
    scat_cols,  scat_rows  = blocks.get("sub_categories",            ([], []))
    var_cols,   var_rows   = blocks.get("variations",                ([], []))
    varo_cols,  varo_rows  = blocks.get("variation_options",         ([], []))
    sta_cols,   sta_rows   = blocks.get("sub_type_attributes",       ([], []))
    stao_cols,  stao_rows  = blocks.get("sub_type_attribute_options",([], []))
    brnd_cols,  brnd_rows  = blocks.get("brands",                    ([], []))

    # --- Build ID maps ---
    dept_map  = build_id_map(dept_rows,  dept_cols.index("id")  if dept_cols  else 0, "dept")
    cat_map   = build_id_map(cat_rows,   cat_cols.index("id")   if cat_cols   else 0, "cat")
    scat_map  = build_id_map(scat_rows,  scat_cols.index("id")  if scat_cols  else 0, "scat")
    var_map   = build_id_map(var_rows,   var_cols.index("id")   if var_cols   else 0, "var")
    varo_map  = build_id_map(varo_rows,  varo_cols.index("id")  if varo_cols  else 0, "varo")
    sta_map   = build_id_map(sta_rows,   sta_cols.index("id")   if sta_cols   else 0, "sta")
    stao_map  = build_id_map(stao_rows,  stao_cols.index("id")  if stao_cols  else 0, "stao")
    brnd_map  = build_id_map(brnd_rows,  brnd_cols.index("id")  if brnd_cols  else 0, "brnd")

    print(
        f"  departments={len(dept_rows)}, categories={len(cat_rows)}, "
        f"sub_categories={len(scat_rows)}\n"
        f"  variations={len(var_rows)}, variation_options={len(varo_rows)}\n"
        f"  sub_type_attributes={len(sta_rows)}, "
        f"sub_type_attribute_options={len(stao_rows)}\n"
        f"  brands={len(brnd_rows)}"
    )

    with open(OUTPUT, "w", encoding="utf-8") as out:
        out.write("-- baseline_seed.sql\n")
        out.write("-- Reference data only. Apply AFTER 'make run' has run migrations.\n")
        out.write("-- Generated from mandi_backup_20260604_074125.sql\n")
        out.write("-- ID format: <prefix>_NNNNN (zero-padded legacy numeric ID)\n\n")

        # DEPARTMENTS — drop: slug
        out.write("-- ------------------------------------------------------------\n")
        out.write("-- departments\n")
        out.write("-- ------------------------------------------------------------\n")
        new_cols = ["id", "name", "sort_order", "is_active", "image_url", "icon"]
        new_rows = [
            [
                dept_map[dict(zip(dept_cols, r))["id"]],
                dict(zip(dept_cols, r))["name"],
                dict(zip(dept_cols, r))["sort_order"],
                dict(zip(dept_cols, r))["is_active"],
                dict(zip(dept_cols, r))["image_url"],
                dict(zip(dept_cols, r))["icon"],
            ]
            for r in dept_rows
        ]
        write_inserts(out, "departments", new_cols, new_rows)

        # CATEGORIES — drop: slug, category_id; remap: department_id
        out.write("-- ------------------------------------------------------------\n")
        out.write("-- categories\n")
        out.write("-- ------------------------------------------------------------\n")
        new_cols = ["id", "department_id", "name", "sort_order", "is_active", "image_url", "icon"]
        new_rows = [
            [
                cat_map[dict(zip(cat_cols, r))["id"]],
                dept_map.get(dict(zip(cat_cols, r))["department_id"], dict(zip(cat_cols, r))["department_id"]),
                dict(zip(cat_cols, r))["name"],
                dict(zip(cat_cols, r))["sort_order"],
                dict(zip(cat_cols, r))["is_active"],
                dict(zip(cat_cols, r))["image_url"],
                dict(zip(cat_cols, r)).get("icon", "\\N"),
            ]
            for r in cat_rows
        ]
        write_inserts(out, "categories", new_cols, new_rows)

        # SUB_CATEGORIES — add: icon=NULL; remap: department_id, category_id
        out.write("-- ------------------------------------------------------------\n")
        out.write("-- sub_categories\n")
        out.write("-- ------------------------------------------------------------\n")
        new_cols = ["id", "department_id", "category_id", "name", "sort_order", "is_active", "image_url", "icon"]
        new_rows = [
            [
                scat_map[dict(zip(scat_cols, r))["id"]],
                dept_map.get(dict(zip(scat_cols, r))["department_id"], dict(zip(scat_cols, r))["department_id"]),
                cat_map.get(dict(zip(scat_cols, r))["category_id"], dict(zip(scat_cols, r))["category_id"]),
                dict(zip(scat_cols, r))["name"],
                dict(zip(scat_cols, r))["sort_order"],
                dict(zip(scat_cols, r))["is_active"],
                dict(zip(scat_cols, r))["image_url"],
                "\\N",  # icon not present in legacy backup; nullable in new schema
            ]
            for r in scat_rows
        ]
        write_inserts(out, "sub_categories", new_cols, new_rows)

        # VARIATIONS — remap: sub_category_id
        out.write("-- ------------------------------------------------------------\n")
        out.write("-- variations\n")
        out.write("-- ------------------------------------------------------------\n")
        new_cols = ["id", "sub_category_id", "name", "sort_order", "is_active"]
        new_rows = [
            [
                var_map[dict(zip(var_cols, r))["id"]],
                scat_map.get(dict(zip(var_cols, r))["sub_category_id"], dict(zip(var_cols, r))["sub_category_id"]),
                dict(zip(var_cols, r))["name"],
                dict(zip(var_cols, r))["sort_order"],
                dict(zip(var_cols, r))["is_active"],
            ]
            for r in var_rows
        ]
        write_inserts(out, "variations", new_cols, new_rows)

        # VARIATION_OPTIONS — remap: variation_id
        out.write("-- ------------------------------------------------------------\n")
        out.write("-- variation_options\n")
        out.write("-- ------------------------------------------------------------\n")
        new_cols = ["id", "variation_id", "value", "sort_order", "is_active"]
        new_rows = [
            [
                varo_map[dict(zip(varo_cols, r))["id"]],
                var_map.get(dict(zip(varo_cols, r))["variation_id"], dict(zip(varo_cols, r))["variation_id"]),
                dict(zip(varo_cols, r))["value"],
                dict(zip(varo_cols, r))["sort_order"],
                dict(zip(varo_cols, r))["is_active"],
            ]
            for r in varo_rows
        ]
        write_inserts(out, "variation_options", new_cols, new_rows)

        # SUB_TYPE_ATTRIBUTES — remap: sub_category_id
        out.write("-- ------------------------------------------------------------\n")
        out.write("-- sub_type_attributes\n")
        out.write("-- ------------------------------------------------------------\n")
        new_cols = ["id", "sub_category_id", "field_name", "field_type", "is_required", "sort_order"]
        new_rows = [
            [
                sta_map[dict(zip(sta_cols, r))["id"]],
                scat_map.get(dict(zip(sta_cols, r))["sub_category_id"], dict(zip(sta_cols, r))["sub_category_id"]),
                dict(zip(sta_cols, r))["field_name"],
                dict(zip(sta_cols, r))["field_type"],
                dict(zip(sta_cols, r))["is_required"],
                dict(zip(sta_cols, r))["sort_order"],
            ]
            for r in sta_rows
        ]
        write_inserts(out, "sub_type_attributes", new_cols, new_rows)

        # SUB_TYPE_ATTRIBUTE_OPTIONS — remap: sub_type_attribute_id
        out.write("-- ------------------------------------------------------------\n")
        out.write("-- sub_type_attribute_options\n")
        out.write("-- ------------------------------------------------------------\n")
        new_cols = ["id", "sub_type_attribute_id", "option_value", "sort_order"]
        new_rows = [
            [
                stao_map[dict(zip(stao_cols, r))["id"]],
                sta_map.get(dict(zip(stao_cols, r))["sub_type_attribute_id"], dict(zip(stao_cols, r))["sub_type_attribute_id"]),
                dict(zip(stao_cols, r))["option_value"],
                dict(zip(stao_cols, r))["sort_order"],
            ]
            for r in stao_rows
        ]
        write_inserts(out, "sub_type_attribute_options", new_cols, new_rows)

        # BRANDS — drop: slug
        out.write("-- ------------------------------------------------------------\n")
        out.write("-- brands\n")
        out.write("-- ------------------------------------------------------------\n")
        new_cols = ["id", "name", "sort_order", "is_active"]
        new_rows = [
            [
                brnd_map[dict(zip(brnd_cols, r))["id"]],
                dict(zip(brnd_cols, r))["name"],
                dict(zip(brnd_cols, r))["sort_order"],
                dict(zip(brnd_cols, r))["is_active"],
            ]
            for r in brnd_rows
        ]
        write_inserts(out, "brands", new_cols, new_rows)

    print(f"Written to {OUTPUT}")


if __name__ == "__main__":
    main()
```

- [ ] **Step 4: Run tests — expect all pass**

```bash
cd mandi-backend
python3 scripts/test_generate_seed.py
```

Expected output:
```
  PASS  test_make_id_pads_to_5_digits
  PASS  test_sql_val_null
  PASS  test_sql_val_booleans
  PASS  test_sql_val_string_escapes_quotes
  PASS  test_sql_val_plain_string
  PASS  test_sql_val_integer_string
  PASS  test_parse_copy_block_basic
  PASS  test_parse_copy_block_empty
  PASS  test_build_id_map

9 passed, 0 failed
```

If any test fails, fix the corresponding function before continuing.

- [ ] **Step 5: Commit**

```bash
git add scripts/generate_baseline_seed.py scripts/test_generate_seed.py
git commit -m "feat: add baseline seed generation script with tests"
```

---

## Task 3: Generate baseline_seed.sql

**Files:**
- Run: `scripts/generate_baseline_seed.py`
- Creates: `baseline_seed.sql`

- [ ] **Step 1: Run the script**

```bash
cd mandi-backend
python3 scripts/generate_baseline_seed.py
```

Expected output (approximate row counts from backup):
```
Reading /Users/harsha/Developer/Localzar/mandi-backend/mandi_backup_20260604_074125.sql ...
  departments=16, categories=170, sub_categories=~700
  variations=0, variation_options=0
  sub_type_attributes=~2570, sub_type_attribute_options=~16194
  brands=0
Written to /Users/harsha/Developer/Localzar/mandi-backend/baseline_seed.sql
```

- [ ] **Step 2: Sanity-check the output**

```bash
# Count INSERT blocks
grep -c "^INSERT INTO" baseline_seed.sql

# Check IDs look right
head -5 baseline_seed.sql
grep "dept_00001" baseline_seed.sql | head -2
grep "cat_00002" baseline_seed.sql | head -2
grep "scat_00016" baseline_seed.sql | head -2

# Check no raw numeric IDs leaked through as FKs
# (all FK values should start with a known prefix)
grep "department_id" baseline_seed.sql | grep -v "dept_" | head -5
grep "category_id" baseline_seed.sql | grep "scat_\|sub_cat" | grep -v "cat_" | head -5
```

Expected for last two commands: **no output** (all FKs remapped).

- [ ] **Step 3: Check file size is reasonable**

```bash
wc -l baseline_seed.sql
```

Expected: somewhere between 20,000 and 25,000 lines (driven by sub_type_attribute_options).

- [ ] **Step 4: Commit the generated seed**

```bash
git add baseline_seed.sql
git commit -m "data: add baseline reference data seed (departments, categories, sub-type attrs)"
```

---

## Task 4: Drop and recreate the database

**Prerequisites:** Docker Desktop must be running before this task.

- [ ] **Step 1: Start Docker Desktop**

Open Docker Desktop application manually. Wait until the whale icon in the menu bar stops animating (Docker is ready).

- [ ] **Step 2: Start PostgreSQL container**

```bash
cd mandi-backend
docker-compose up -d postgresdb
```

Expected output:
```
[+] Running 1/1
 ✔ Container mandi-backend-postgresdb-1  Started
```

- [ ] **Step 3: Wait for Postgres to be ready**

```bash
sleep 3
docker-compose exec postgresdb pg_isready -U rohitjangid
```

Expected: `/var/run/postgresql:5432 - accepting connections`

If not ready, wait 5 more seconds and retry.

- [ ] **Step 4: Drop and recreate the mandi database**

```bash
docker-compose exec postgresdb psql -U rohitjangid -d postgres \
  -c "DROP DATABASE IF EXISTS mandi;" \
  -c "CREATE DATABASE mandi OWNER rohitjangid;"
```

Expected:
```
DROP DATABASE
CREATE DATABASE
```

- [ ] **Step 5: Verify the database is empty**

```bash
docker-compose exec postgresdb psql -U rohitjangid -d mandi \
  -c "\dt"
```

Expected: `Did not find any relations.`

---

## Task 5: Run migrations via make run

**Files:** No changes — migrations run automatically on startup.

- [ ] **Step 1: Run the backend (it will apply all migrations then start)**

```bash
cd mandi-backend
make run
```

Watch the logs. Expected sequence (within first 10 seconds):
```
... running migrations ...
... migration 000001 applied ...
... migration 000002 applied (or skipped — no change) ...
...
... migration 000007 applied ...
... saving payment methods ...
... seeding countries ...
... server started on :3000 ...
```

If you see `database migration failed:` — stop, read the error, and check Step 2.

- [ ] **Step 2: Confirm all 7 migrations applied**

While the server is running, open a second terminal:

```bash
cd mandi-backend
docker-compose exec postgresdb psql -U rohitjangid -d mandi \
  -c "SELECT version, dirty FROM schema_migrations ORDER BY version;"
```

Expected:
```
 version | dirty
---------+-------
       1 | f
       2 | f
       3 | f
       4 | f
       5 | f
       6 | f
       7 | f
(7 rows)
```

If any row has `dirty = t`, the migration failed midway — check the server logs for the error.

- [ ] **Step 3: Confirm key tables exist**

```bash
docker-compose exec postgresdb psql -U rohitjangid -d mandi \
  -c "\dt" | grep -E "departments|categories|sub_categories|admins|users|shop_details"
```

Expected: all 6 table names appear in output.

- [ ] **Step 4: Confirm payment_methods and countries were seeded**

```bash
docker-compose exec postgresdb psql -U rohitjangid -d mandi \
  -c "SELECT name FROM payment_methods;" \
  -c "SELECT count(*) FROM countries;"
```

Expected:
```
    name
-----------
 cod
 razor pay
 stripe

 count
-------
    10
```

- [ ] **Step 5: Stop the server**

Press `Ctrl+C` in the terminal running `make run`.

---

## Task 6: Apply the baseline seed

- [ ] **Step 1: Apply baseline_seed.sql via Docker psql**

```bash
cd mandi-backend
docker-compose exec -T postgresdb psql -U rohitjangid -d mandi < baseline_seed.sql
```

The `-T` flag disables pseudo-TTY allocation so stdin piping works. This will take 20–60 seconds (large sub_type_attribute_options insert).

Expected: No error lines. The command exits with code 0.

If you see `ERROR: insert or update on table "categories" violates foreign key constraint`, the departments insert failed — check the departments block in `baseline_seed.sql` first.

- [ ] **Step 2: Verify row counts**

```bash
docker-compose exec postgresdb psql -U rohitjangid -d mandi -c "
SELECT 'departments' AS tbl, count(*) FROM departments
UNION ALL SELECT 'categories', count(*) FROM categories
UNION ALL SELECT 'sub_categories', count(*) FROM sub_categories
UNION ALL SELECT 'sub_type_attributes', count(*) FROM sub_type_attributes
UNION ALL SELECT 'sub_type_attribute_options', count(*) FROM sub_type_attribute_options
UNION ALL SELECT 'brands', count(*) FROM brands
ORDER BY tbl;
"
```

Expected approximate counts (based on backup):
```
           tbl            | count
--------------------------+-------
 brands                   |     0
 categories               |   170+
 departments              |    16
 sub_categories           |   700+
 sub_type_attribute_options| 16000+
 sub_type_attributes      |  2500+
```

- [ ] **Step 3: Verify FK integrity with a spot-check**

```bash
docker-compose exec postgresdb psql -U rohitjangid -d mandi -c "
SELECT d.name AS dept, c.name AS cat, sc.name AS subcat
FROM sub_categories sc
JOIN categories c ON sc.category_id = c.id
JOIN departments d ON sc.department_id = d.id
LIMIT 10;
"
```

Expected: 10 rows with readable department → category → sub-category chains. No FK errors.

---

## Task 7: Smoke-test the running backend

- [ ] **Step 1: Start the server**

```bash
cd mandi-backend
make run
```

Wait for `server started on :3000` in logs.

- [ ] **Step 2: Health check**

```bash
curl -s http://localhost:3000/api/health
```

Expected: `{"status":"ok"}`

- [ ] **Step 3: Check categories endpoint**

```bash
curl -s "http://localhost:3000/api/categories?limit=5&offset=0" | python3 -m json.tool | head -30
```

Expected: JSON response with `"success": true` and a `data` array containing category objects with `category_name`, `department_id` fields populated.

- [ ] **Step 4: Check departments endpoint**

```bash
curl -s "http://localhost:3000/api/departments?limit=5&offset=0" | python3 -m json.tool | head -20
```

Expected: `"success": true` with 5 department objects.

- [ ] **Step 5: Check server logs for errors**

In the terminal running `make run`, scan for any `ERROR` or `panic` lines after startup. There should be none during the health/category/department requests.

- [ ] **Step 6: Final commit**

```bash
# Stop the server with Ctrl+C first, then:
git add -A
git status  # review — should only show baseline_seed.sql if not already committed
git commit -m "chore: verify clean DB restore — migrations healthy, seed applied"
```

---

## Self-Review Checklist

**Spec coverage:**
- ✅ Python seed generation script → Task 2 + 3
- ✅ Migration verification → Task 1
- ✅ Drop/recreate DB → Task 4
- ✅ Run migrations (make run) → Task 5
- ✅ Apply baseline_seed.sql → Task 6
- ✅ Smoke-test → Task 7
- ✅ All schema mismatch columns handled in the script (slug dropped, category_id dropped, icon added as NULL)

**No placeholders:** All steps have exact commands and expected output.

**Type consistency:** `make_id`, `sql_val`, `parse_copy_blocks_from_string`, `build_id_map`, `write_inserts` are defined in Task 2 Step 3 and referenced consistently in tests (Task 2 Step 1).
