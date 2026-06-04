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
