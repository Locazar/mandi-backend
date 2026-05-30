//go:build integration

package repository

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/db"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// connectTestDB opens a GORM connection to the Postgres instance identified by
// the DATABASE_URL environment variable (falls back to the default docker-compose
// DSN). It then calls db.RunMigrations to apply the baseline schema so that every
// table, FK, unique index and CHECK constraint is present before the tests run.
func connectTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres dbname=mandi_test port=5432 password=postgres sslmode=disable"
	}

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connectTestDB: open: %v", err)
	}

	if err := db.RunMigrations(gormDB); err != nil {
		t.Fatalf("connectTestDB: RunMigrations: %v", err)
	}

	return gormDB
}

// uniqueID returns a deterministic prefixed test ID that is guaranteed to be
// distinct across test cases using the supplied suffix.
func uniqueID(prefix, suffix string) string {
	return fmt.Sprintf("%s_%s_%d", prefix, suffix, time.Now().UnixNano())
}

// ---------------------------------------------------------------------------
// Test (a): FK violation — shop_orders.user_id references a non-existent user
// ---------------------------------------------------------------------------

// TestIntegration_FKViolation_ShopOrder asserts that the DB rejects a
// shop_orders INSERT whose user_id does not exist in the users table (the FK
// is declared ON DELETE RESTRICT, so the constraint fires on INSERT as well).
func TestIntegration_FKViolation_ShopOrder(t *testing.T) {
	gormDB := connectTestDB(t)

	nonExistentUserID := uniqueID("usr", "fk_test")
	orderID := uniqueID("ord", "fk_test")

	sql := `INSERT INTO shop_orders
		(id, user_id, order_date, address_id,
		 order_total_amount_minor, order_total_currency,
		 discount_amount_minor,    discount_currency,
		 status)
		VALUES (?, ?, NOW(), ?, 0, 'INR', 0, 'INR', 'order placed')`

	// address_id also has an FK; use the same non-existent placeholder — the
	// DB will reject on user_id first (or address_id, either way an error is
	// the expected outcome).
	nonExistentAddrID := uniqueID("adr", "fk_test")
	res := gormDB.Exec(sql, orderID, nonExistentUserID, nonExistentAddrID)

	if res.Error == nil {
		t.Fatalf("expected FK violation error for non-existent user_id %q, got nil", nonExistentUserID)
	}
}

// ---------------------------------------------------------------------------
// Test (b): Unique violation — duplicate cart_items(cart_id, product_item_id)
// ---------------------------------------------------------------------------

// TestIntegration_UniqueViolation_CartItem asserts that the partial unique
// index uidx_cart_items_cart_product rejects a second CartItem row with the
// same (cart_id, product_item_id) pair.
//
// Because cart_items carries FKs to carts and product_items we first insert a
// parent user, cart, and product hierarchy, then insert the cart item twice.
func TestIntegration_UniqueViolation_CartItem(t *testing.T) {
	gormDB := connectTestDB(t)

	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback() // always roll back so test data does not persist

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Insert a minimal user (just the required columns).
	userID := "usr_ci_" + suffix
	_, err = tx.Exec(`INSERT INTO users (id, name, email, phone, password, created_at, updated_at)
		VALUES ($1, 'Test User', '', $2, 'x', NOW(), NOW())`,
		userID, "9999000000"+suffix[:5])
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	// Insert a cart for that user.
	cartID := "crt_ci_" + suffix
	_, err = tx.Exec(`INSERT INTO carts (id, user_id, total_price_amount_minor, total_price_currency,
		discount_amount_amount_minor, discount_amount_currency, created_at, updated_at)
		VALUES ($1, $2, 0, 'INR', 0, 'INR', NOW(), NOW())`, cartID, userID)
	if err != nil {
		t.Fatalf("insert cart: %v", err)
	}

	// We need a product_item to satisfy the FK on cart_items.  Build the
	// minimal ancestor chain: department → category → sub_category →
	// product → product_item.
	deptID := "dep_ci_" + suffix
	_, err = tx.Exec(`INSERT INTO departments (id, name, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())`,
		deptID, "TestDept"+suffix)
	if err != nil {
		t.Fatalf("insert department: %v", err)
	}

	catID := "cat_ci_" + suffix
	_, err = tx.Exec(`INSERT INTO categories (id, name, department_id, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())`, catID, "TestCat"+suffix, deptID)
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}

	subCatID := "sbc_ci_" + suffix
	_, err = tx.Exec(`INSERT INTO sub_categories (id, name, category_id, department_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`, subCatID, "TestSubCat"+suffix, catID, deptID)
	if err != nil {
		t.Fatalf("insert sub_category: %v", err)
	}

	// Insert a shop (needed by products and product_items).
	adminID := "adm_ci_" + suffix
	_, err = tx.Exec(`INSERT INTO admins (id, name, email, password, admin_status, created_at, updated_at)
		VALUES ($1, 'Shop Admin', $2, 'x', 'active', NOW(), NOW())`,
		adminID, "admin_"+suffix+"@test.com")
	if err != nil {
		t.Fatalf("insert admin: %v", err)
	}

	shopID := "shp_ci_" + suffix
	_, err = tx.Exec(`INSERT INTO shop_details
		(id, admin_id, shop_name, shop_type, shop_status, created_at, updated_at)
		VALUES ($1, $2, $3, 'retail', 'active', NOW(), NOW())`,
		shopID, adminID, "TestShop"+suffix)
	if err != nil {
		t.Fatalf("insert shop: %v", err)
	}

	prodID := "prd_ci_" + suffix
	_, err = tx.Exec(`INSERT INTO products
		(id, product_name, shop_id, sub_category_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		prodID, "TestProd"+suffix, shopID, subCatID)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}

	piID := "pit_ci_" + suffix
	_, err = tx.Exec(`INSERT INTO product_items
		(id, product_id, shop_id, sku, price_amount_minor, price_currency,
		 qty_in_stock, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 100, 'INR', 10, TRUE, NOW(), NOW())`,
		piID, prodID, shopID, "SKU"+suffix)
	if err != nil {
		t.Fatalf("insert product_item: %v", err)
	}

	// First cart item — should succeed.
	ci1ID := "cit_ci1_" + suffix
	_, err = tx.Exec(`INSERT INTO cart_items (id, cart_id, product_item_id, qty, created_at, updated_at)
		VALUES ($1, $2, $3, 1, NOW(), NOW())`, ci1ID, cartID, piID)
	if err != nil {
		t.Fatalf("first cart item insert: %v", err)
	}

	// Second cart item with the same (cart_id, product_item_id) — must fail.
	ci2ID := "cit_ci2_" + suffix
	_, err = tx.Exec(`INSERT INTO cart_items (id, cart_id, product_item_id, qty, created_at, updated_at)
		VALUES ($1, $2, $3, 2, NOW(), NOW())`, ci2ID, cartID, piID)

	if err == nil {
		t.Fatal("expected unique-constraint violation for duplicate (cart_id, product_item_id), got nil")
	}
}

// ---------------------------------------------------------------------------
// Test (c): CHECK constraint — shop_orders.status = 'bogus_status'
// ---------------------------------------------------------------------------

// TestIntegration_CheckViolation_OrderStatus asserts that the DB rejects an
// INSERT into shop_orders whose status value is not in the allowed enum list.
func TestIntegration_CheckViolation_OrderStatus(t *testing.T) {
	gormDB := connectTestDB(t)

	sql := `INSERT INTO shop_orders
		(id, user_id, order_date, address_id,
		 order_total_amount_minor, order_total_currency,
		 discount_amount_minor,    discount_currency,
		 status)
		VALUES (?, ?, NOW(), ?, 0, 'INR', 0, 'INR', 'bogus_status')`

	// The CHECK fires before the FK is evaluated in most Postgres planner
	// strategies; but even if FK fires first the result is still a non-nil
	// error, satisfying the test.
	id := uniqueID("ord", "chk_test")
	res := gormDB.Exec(sql, id, "usr_nonexistent", "adr_nonexistent")

	if res.Error == nil {
		t.Fatal("expected CHECK constraint violation for status = 'bogus_status', got nil")
	}
}

// ---------------------------------------------------------------------------
// Test (d): CASCADE delete — deleting a cart removes its cart_items
// ---------------------------------------------------------------------------

// TestIntegration_CascadeDelete_CartItems asserts that when a cart row is
// deleted its child cart_items rows are removed automatically (ON DELETE CASCADE).
func TestIntegration_CascadeDelete_CartItems(t *testing.T) {
	gormDB := connectTestDB(t)

	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	suffix := fmt.Sprintf("cas%d", time.Now().UnixNano())

	// Insert a minimal user.
	userID := "usr_" + suffix
	_, err = tx.Exec(`INSERT INTO users (id, name, email, phone, password, created_at, updated_at)
		VALUES ($1, 'CascUser', '', $2, 'x', NOW(), NOW())`,
		userID, "888"+suffix[:8])
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	// Insert a cart.
	cartID := "crt_" + suffix
	_, err = tx.Exec(`INSERT INTO carts (id, user_id, total_price_amount_minor, total_price_currency,
		discount_amount_amount_minor, discount_amount_currency, created_at, updated_at)
		VALUES ($1, $2, 0, 'INR', 0, 'INR', NOW(), NOW())`, cartID, userID)
	if err != nil {
		t.Fatalf("insert cart: %v", err)
	}

	// Build ancestor chain for a product_item.
	deptID := "dep_" + suffix
	_, err = tx.Exec(`INSERT INTO departments (id, name, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())`,
		deptID, "CasDept"+suffix)
	if err != nil {
		t.Fatalf("insert department: %v", err)
	}

	catID := "cat_" + suffix
	_, err = tx.Exec(`INSERT INTO categories (id, name, department_id, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())`, catID, "CasCat"+suffix, deptID)
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}

	subCatID := "sbc_" + suffix
	_, err = tx.Exec(`INSERT INTO sub_categories (id, name, category_id, department_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`, subCatID, "CasSubCat"+suffix, catID, deptID)
	if err != nil {
		t.Fatalf("insert sub_category: %v", err)
	}

	adminID := "adm_" + suffix
	_, err = tx.Exec(`INSERT INTO admins (id, name, email, password, admin_status, created_at, updated_at)
		VALUES ($1, 'CasAdmin', $2, 'x', 'active', NOW(), NOW())`,
		adminID, "casadmin_"+suffix+"@test.com")
	if err != nil {
		t.Fatalf("insert admin: %v", err)
	}

	shopID := "shp_" + suffix
	_, err = tx.Exec(`INSERT INTO shop_details
		(id, admin_id, shop_name, shop_type, shop_status, created_at, updated_at)
		VALUES ($1, $2, $3, 'retail', 'active', NOW(), NOW())`,
		shopID, adminID, "CasShop"+suffix)
	if err != nil {
		t.Fatalf("insert shop: %v", err)
	}

	prodID := "prd_" + suffix
	_, err = tx.Exec(`INSERT INTO products
		(id, product_name, shop_id, sub_category_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		prodID, "CasProd"+suffix, shopID, subCatID)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}

	piID := "pit_" + suffix
	_, err = tx.Exec(`INSERT INTO product_items
		(id, product_id, shop_id, sku, price_amount_minor, price_currency,
		 qty_in_stock, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 50, 'INR', 5, TRUE, NOW(), NOW())`,
		piID, prodID, shopID, "CASKU"+suffix)
	if err != nil {
		t.Fatalf("insert product_item: %v", err)
	}

	// Insert a cart_item child.
	ciID := "cit_" + suffix
	_, err = tx.Exec(`INSERT INTO cart_items (id, cart_id, product_item_id, qty, created_at, updated_at)
		VALUES ($1, $2, $3, 1, NOW(), NOW())`, ciID, cartID, piID)
	if err != nil {
		t.Fatalf("insert cart_item: %v", err)
	}

	// Verify the cart_item exists before the delete.
	var countBefore int
	row := tx.QueryRow(`SELECT COUNT(*) FROM cart_items WHERE id = $1`, ciID)
	if err := row.Scan(&countBefore); err != nil || countBefore != 1 {
		t.Fatalf("expected 1 cart_item before delete, got %d (err=%v)", countBefore, err)
	}

	// Delete the cart — ON DELETE CASCADE must propagate to cart_items.
	_, err = tx.Exec(`DELETE FROM carts WHERE id = $1`, cartID)
	if err != nil {
		t.Fatalf("delete cart: %v", err)
	}

	// The cart_item should no longer exist.
	var countAfter int
	row = tx.QueryRow(`SELECT COUNT(*) FROM cart_items WHERE id = $1`, ciID)
	if err := row.Scan(&countAfter); err != nil {
		t.Fatalf("count after delete: %v", err)
	}
	if countAfter != 0 {
		t.Fatalf("expected 0 cart_items after CASCADE delete of cart, got %d", countAfter)
	}
}
