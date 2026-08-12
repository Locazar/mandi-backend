package repository

import (
	"context"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// Reserved customer-only "All" department resolution.
//
// An admin creates a normal department named "All" (see domain.AllDepartmentName)
// via the admin-portal. Customer queries treat that department's id as "no
// department filter" so the customer app sees every department's products,
// shops and categories through a single entry — with no customer-app change.
//
// The id is resolved by name once and cached for a short window so the hot
// customer paths (product search, shop list, categories) don't hit the DB on
// every request. When no such department exists the cached id is empty and the
// feature is simply off; when an admin later creates/deletes it, the change
// takes effect within the cache TTL.

const allDeptCacheTTL = 60 * time.Second

var (
	allDeptMu       sync.RWMutex
	allDeptCachedID string
	allDeptCachedAt time.Time
)

// AllDepartmentID returns the id of the reserved "All" department (matched by
// name, case-insensitive), cached briefly. Returns "" when no such department
// exists — i.e. the feature is off until an admin creates it.
func AllDepartmentID(ctx context.Context, db *gorm.DB) string {
	allDeptMu.RLock()
	fresh := !allDeptCachedAt.IsZero() && time.Since(allDeptCachedAt) < allDeptCacheTTL
	cached := allDeptCachedID
	allDeptMu.RUnlock()
	if fresh {
		return cached
	}

	var found string
	if db != nil {
		// Best-effort: on error we cache "" (feature off) and retry after the TTL.
		_ = db.WithContext(ctx).
			Table("departments").
			Select("id").
			Where("LOWER(name) = ?", strings.ToLower(domain.AllDepartmentName)).
			Limit(1).
			Scan(&found).Error
	}

	allDeptMu.Lock()
	allDeptCachedID = found
	allDeptCachedAt = time.Now()
	allDeptMu.Unlock()
	return found
}

// IsAllDepartment reports whether departmentID is the reserved "All" department.
func IsAllDepartment(ctx context.Context, db *gorm.DB, departmentID string) bool {
	if strings.TrimSpace(departmentID) == "" {
		return false
	}
	allID := AllDepartmentID(ctx, db)
	return allID != "" && allID == departmentID
}
