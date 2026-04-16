package db

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// SetupSearchIndexes installs pg_trgm and creates trigram GIN indexes for
// global search and autocomplete.  Safe to call on every startup — all
// statements are idempotent (IF NOT EXISTS / IF NOT EXISTS on the extension).
func SetupSearchIndexes(db *gorm.DB) error {
	// Ensure the extension exists regardless of how Postgres was initialised
	// (Docker init scripts only run once on a fresh volume; make run against
	// an existing DB may never have run them).
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS pg_trgm`).Error; err != nil {
		log.Printf("Warning: failed to create pg_trgm extension: %v", err)
		return nil // cannot create trigram indexes without the extension
	}

	// table → index DDL pairs; skip the index if the table does not yet exist
	type indexDef struct {
		table string
		ddl   string
	}

	defs := []indexDef{
		{
			table: "products",
			ddl:   `CREATE INDEX IF NOT EXISTS idx_products_name_trgm ON products USING gin (name gin_trgm_ops)`,
		},
		{
			table: "products",
			ddl:   `CREATE INDEX IF NOT EXISTS idx_products_fts ON products USING gin (to_tsvector('english', COALESCE(name, '') || ' ' || COALESCE(description, '')))`,
		},
		{
			table: "categories",
			ddl:   `CREATE INDEX IF NOT EXISTS idx_categories_name_trgm ON categories USING gin (name gin_trgm_ops)`,
		},
		{
			table: "shop_details",
			ddl:   `CREATE INDEX IF NOT EXISTS idx_shop_details_shop_name_trgm ON shop_details USING gin (shop_name gin_trgm_ops)`,
		},
		{
			table: "brands",
			ddl:   `CREATE INDEX IF NOT EXISTS idx_brands_name_trgm ON brands USING gin (name gin_trgm_ops)`,
		},
		{
			table: "departments",
			ddl:   `CREATE INDEX IF NOT EXISTS idx_departments_name_trgm ON departments USING gin (name gin_trgm_ops)`,
		},
		{
			table: "offers",
			ddl:   `CREATE INDEX IF NOT EXISTS idx_offers_name_trgm ON offers USING gin (name gin_trgm_ops)`,
		},
		{
			table: "offers",
			ddl:   `CREATE INDEX IF NOT EXISTS idx_offers_description_trgm ON offers USING gin (description gin_trgm_ops)`,
		},
	}

	for _, d := range defs {
		if !tableExists(db, d.table) {
			log.Printf("Search index skipped: table %q does not exist yet", d.table)
			continue
		}
		if err := db.Exec(d.ddl).Error; err != nil {
			log.Printf("Warning: failed to create search index on %q: %v", d.table, err)
		}
	}

	log.Println("successfully set up search indexes")
	return nil
}

// tableExists returns true when the named table is present in the public schema.
func tableExists(db *gorm.DB, table string) bool {
	var count int64
	db.Raw(
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?`,
		table,
	).Scan(&count)
	return count > 0
}

// EnsureSearchExtension installs pg_trgm without touching indexes.
// Exported so callers who only need the extension can use it directly.
func EnsureSearchExtension(db *gorm.DB) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS pg_trgm`).Error; err != nil {
		return fmt.Errorf("pg_trgm extension: %w", err)
	}
	return nil
}
