package cloud

import "strings"

// ResolveURL converts a stored image-column value to a URL ready for clients.
//
// Three input shapes exist during the migration window:
//   - empty                                            → ""
//   - full http(s)://…                                  → unchanged (e.g., pre-existing S3 user-profile rows)
//   - legacy "uploads/<dir>/<file>" or "/uploads/…"    → "/uploads/<dir>/<file>" served via StaticFS
//   - bare object key (no scheme, no "uploads/" prefix)→ cs.PublicURL(key)
//
// The "uploads/..." branch is removed in a later phase once the backfill is complete.
func ResolveURL(cs CloudService, stored string) string {
	if stored == "" {
		return ""
	}
	if strings.HasPrefix(stored, "http://") || strings.HasPrefix(stored, "https://") {
		return stored
	}
	if strings.HasPrefix(stored, "uploads/") || strings.HasPrefix(stored, "/uploads/") {
		if !strings.HasPrefix(stored, "/") {
			return "/" + stored
		}
		return stored
	}
	return cs.PublicURL(stored)
}
