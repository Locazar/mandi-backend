// migrate-uploads is a one-time backfill that copies media from ./uploads/
// into the configured object-storage bucket and rewrites DB columns to bare
// object keys.
//
// Idempotent: rows whose value is empty, already an http(s) URL, or already
// look like a bare key (no "uploads/" prefix) are skipped. Safe to re-run.
//
// Usage:
//   go run ./cmd/migrate-uploads --dry-run
//   go run ./cmd/migrate-uploads --limit 5 --namespace products
//   go run ./cmd/migrate-uploads                 # full run
//   go run ./cmd/migrate-uploads --uploads-dir /path/to/uploads
//
// Target columns (app-authoritative per Phase 0 trace):
//   admins.profile_image_url           -> admin-profiles/
//   products.image                     -> products/
//   product_items.product_item_images  -> products/        (text[] array)
//   categories.image_url               -> category-images/
//   departments.image_url              -> departments/
//   sub_categories.image_url           -> sub-category-images/
//   banners.image_url                  -> banners/
//   offers.image                       -> offers/
//   offers.thumbnail                   -> offers/thumbnail/
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
)

type target struct {
	table     string
	column    string
	isArray   bool
	namespace string
}

var targets = []target{
	{table: "admins", column: "profile_image_url", namespace: "admin-profiles"},
	{table: "products", column: "image", namespace: "products"},
	{table: "product_items", column: "product_item_images", isArray: true, namespace: "products"},
	{table: "categories", column: "image_url", namespace: "category-images"},
	{table: "departments", column: "image_url", namespace: "departments"},
	{table: "sub_categories", column: "image_url", namespace: "sub-category-images"},
	{table: "banners", column: "image_url", namespace: "banners"},
	{table: "offers", column: "image", namespace: "offers"},
	{table: "offers", column: "thumbnail", namespace: "offers/thumbnail"},
}

type stats struct {
	planned   int
	uploaded  int
	skipped   int
	missing   int
	updateErr int
}

func (s *stats) add(o stats) {
	s.planned += o.planned
	s.uploaded += o.uploaded
	s.skipped += o.skipped
	s.missing += o.missing
	s.updateErr += o.updateErr
}

func main() {
	dryRun := flag.Bool("dry-run", false, "scan only; do not upload or update DB")
	limit := flag.Int("limit", 0, "max rows per target (0 = no limit)")
	onlyNamespace := flag.String("namespace", "", "restrict to this namespace")
	uploadsDir := flag.String("uploads-dir", "./uploads", "root of local upload tree")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	gdb, err := db.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	cs, err := cloud.NewObjectStorageService(cfg)
	if err != nil {
		log.Fatalf("init cloud: %v", err)
	}

	ctx := context.Background()
	total := stats{}
	mode := "LIVE"
	if *dryRun {
		mode = "DRY-RUN"
	}
	log.Printf("=== backfill mode=%s uploads-dir=%s limit=%d namespace=%q ===",
		mode, *uploadsDir, *limit, *onlyNamespace)

	for _, t := range targets {
		if *onlyNamespace != "" && t.namespace != *onlyNamespace {
			continue
		}
		s, err := migrateTarget(ctx, gdb, cs, t, *uploadsDir, *limit, *dryRun)
		total.add(s)
		if err != nil {
			log.Printf("[%s.%s] aborted with error: %v", t.table, t.column, err)
		}
		log.Printf("[%s.%s] planned=%d uploaded=%d skipped=%d missing-file=%d update-err=%d",
			t.table, t.column, s.planned, s.uploaded, s.skipped, s.missing, s.updateErr)
	}

	log.Printf("=== TOTAL planned=%d uploaded=%d skipped=%d missing-file=%d update-err=%d (mode=%s) ===",
		total.planned, total.uploaded, total.skipped, total.missing, total.updateErr, mode)
}

func migrateTarget(ctx context.Context, gdb *gorm.DB, cs cloud.CloudService, t target, uploadsDir string, limit int, dryRun bool) (stats, error) {
	s := stats{}

	q := fmt.Sprintf("SELECT id, %s FROM %s WHERE %s IS NOT NULL ORDER BY id", t.column, t.table, t.column)
	if limit > 0 {
		q = fmt.Sprintf("%s LIMIT %d", q, limit)
	}
	rows, err := gdb.Raw(q).Rows()
	if err != nil {
		return s, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uint
		var raw string
		if err := rows.Scan(&id, &raw); err != nil {
			return s, fmt.Errorf("scan: %w", err)
		}

		if t.isArray {
			s2 := migrateArrayRow(ctx, gdb, cs, t, uploadsDir, id, raw, dryRun)
			s.add(s2)
			continue
		}

		action := classify(raw)
		switch action {
		case actionSkip:
			s.skipped++
		case actionMigrate:
			s.planned++
			if dryRun {
				continue
			}
			newKey, err := uploadFromDisk(ctx, cs, raw, t.namespace, uploadsDir)
			if errors.Is(err, errFileMissing) {
				s.missing++
				log.Printf("[%s.%s] id=%d missing file %q", t.table, t.column, id, raw)
				continue
			}
			if err != nil {
				log.Printf("[%s.%s] id=%d upload error: %v", t.table, t.column, id, err)
				continue
			}
			if err := gdb.Exec(
				fmt.Sprintf("UPDATE %s SET %s = ? WHERE id = ?", t.table, t.column),
				newKey, id,
			).Error; err != nil {
				s.updateErr++
				log.Printf("[%s.%s] id=%d update error: %v", t.table, t.column, id, err)
				continue
			}
			s.uploaded++
		}
	}
	return s, rows.Err()
}

func migrateArrayRow(ctx context.Context, gdb *gorm.DB, cs cloud.CloudService, t target, uploadsDir string, id uint, raw string, dryRun bool) stats {
	s := stats{}
	arr := parsePGTextArray(raw)
	changed := false
	newArr := make([]string, len(arr))
	for i, v := range arr {
		newArr[i] = v
		switch classify(v) {
		case actionSkip:
			s.skipped++
			continue
		case actionMigrate:
			s.planned++
			if dryRun {
				continue
			}
			newKey, err := uploadFromDisk(ctx, cs, v, t.namespace, uploadsDir)
			if errors.Is(err, errFileMissing) {
				s.missing++
				log.Printf("[%s.%s][arr] id=%d missing file %q", t.table, t.column, id, v)
				continue
			}
			if err != nil {
				log.Printf("[%s.%s][arr] id=%d upload error: %v", t.table, t.column, id, err)
				continue
			}
			newArr[i] = newKey
			s.uploaded++
			changed = true
		}
	}
	if !dryRun && changed {
		newRaw := buildPGTextArray(newArr)
		if err := gdb.Exec(
			fmt.Sprintf("UPDATE %s SET %s = ?::text[] WHERE id = ?", t.table, t.column),
			newRaw, id,
		).Error; err != nil {
			s.updateErr++
			log.Printf("[%s.%s] id=%d update error: %v", t.table, t.column, id, err)
		}
	}
	return s
}

type action int

const (
	actionSkip action = iota
	actionMigrate
)

func classify(v string) action {
	v = strings.TrimSpace(v)
	if v == "" {
		return actionSkip
	}
	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		return actionSkip
	}
	if !strings.HasPrefix(v, "uploads/") && !strings.HasPrefix(v, "/uploads/") {
		return actionSkip // already migrated or unknown shape
	}
	return actionMigrate
}

var errFileMissing = errors.New("local file missing")

func uploadFromDisk(ctx context.Context, cs cloud.CloudService, stored, namespace, uploadsDir string) (string, error) {
	// Normalize: strip leading "/" then "uploads/" then any leftover separators.
	rel := strings.TrimPrefix(strings.TrimPrefix(stored, "/"), "uploads/")
	rel = strings.ReplaceAll(rel, "\\", "/") // tolerate windows-style separators
	diskPath := filepath.Join(uploadsDir, rel)

	data, err := os.ReadFile(diskPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errFileMissing
		}
		return "", fmt.Errorf("read %s: %w", diskPath, err)
	}

	filename := filepath.Base(rel)
	ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename)))
	if ct == "" {
		ct = "application/octet-stream"
	}
	key, err := cs.SaveBytes(ctx, data, cloud.SaveOptions{
		Namespace:   namespace,
		Visibility:  cloud.VisibilityPublic,
		ContentType: ct,
		Filename:    filename,
	})
	if err != nil {
		return "", err
	}
	return key, nil
}

// parsePGTextArray decodes PostgreSQL text[] literal "{a,b,c}" into []string.
// Does not handle quoted elements with embedded commas (none exist in our data).
func parsePGTextArray(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		return nil
	}
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		// PostgreSQL may quote strings containing special chars.
		p = strings.TrimPrefix(p, "\"")
		p = strings.TrimSuffix(p, "\"")
		// pgx encodes backslashes; normalize back to forward slashes.
		p = strings.ReplaceAll(p, "\\\\", "/")
		p = strings.ReplaceAll(p, "\\", "/")
		out = append(out, p)
	}
	return out
}

// buildPGTextArray builds a PostgreSQL text[] literal "{a,b,c}" from []string.
func buildPGTextArray(arr []string) string {
	if len(arr) == 0 {
		return "{}"
	}
	return "{" + strings.Join(arr, ",") + "}"
}
