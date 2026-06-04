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
    dept_cols,  dept_rows  = blocks.get("departments",                ([], []))
    cat_cols,   cat_rows   = blocks.get("categories",                 ([], []))
    scat_cols,  scat_rows  = blocks.get("sub_categories",             ([], []))
    var_cols,   var_rows   = blocks.get("variations",                 ([], []))
    varo_cols,  varo_rows  = blocks.get("variation_options",          ([], []))
    sta_cols,   sta_rows   = blocks.get("sub_type_attributes",        ([], []))
    stao_cols,  stao_rows  = blocks.get("sub_type_attribute_options", ([], []))
    brnd_cols,  brnd_rows  = blocks.get("brands",                     ([], []))

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
