package repository

import (
	"context"
	"testing"
	"time"
)

// resetAllDeptCache clears the package-level cache so tests don't leak state.
func resetAllDeptCache() {
	allDeptMu.Lock()
	allDeptCachedID = ""
	allDeptCachedAt = time.Time{}
	allDeptMu.Unlock()
}

func TestIsAllDepartment_Guards(t *testing.T) {
	resetAllDeptCache()
	ctx := context.Background()

	// Empty / whitespace department id is never the "All" department, and must
	// not touch the DB (nil db here would panic if it did).
	if IsAllDepartment(ctx, nil, "") {
		t.Fatal("empty department id must not be treated as All")
	}
	if IsAllDepartment(ctx, nil, "   ") {
		t.Fatal("whitespace department id must not be treated as All")
	}

	// With no DB the feature is off: AllDepartmentID resolves to "" and no real
	// id can match, so the department filter is always applied normally.
	if got := AllDepartmentID(ctx, nil); got != "" {
		t.Fatalf("AllDepartmentID with nil db = %q, want empty", got)
	}
	if IsAllDepartment(ctx, nil, "dept_00001") {
		t.Fatal("a real department id must not match when no All department exists")
	}
}
