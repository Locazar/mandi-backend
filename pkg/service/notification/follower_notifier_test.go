package notification

import (
	"context"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
)

func TestResolveFollowerImageURL(t *testing.T) {
	t.Run("absolute passes through", func(t *testing.T) {
		t.Setenv("NOTIFICATION_PUBLIC_BASE_URL", "https://ignored.example")
		if got := resolveFollowerImageURL(nil, "https://cdn.x/a.jpg"); got != "https://cdn.x/a.jpg" {
			t.Fatalf("absolute url changed: %q", got)
		}
	})

	t.Run("empty stays empty", func(t *testing.T) {
		if got := resolveFollowerImageURL(nil, ""); got != "" {
			t.Fatalf("expected empty, got %q", got)
		}
	})

	t.Run("relative joins the public base", func(t *testing.T) {
		t.Setenv("NOTIFICATION_PUBLIC_BASE_URL", "")
		t.Setenv("PUBLIC_BASE_URL", "https://api.locazar.com/")
		t.Setenv("S3_PUBLIC_BASE_URL", "")
		want := "https://api.locazar.com/uploads/products/x.jpg"
		if got := resolveFollowerImageURL(nil, "uploads/products/x.jpg"); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("relative with no env base falls back to the API host", func(t *testing.T) {
		t.Setenv("NOTIFICATION_PUBLIC_BASE_URL", "")
		t.Setenv("PUBLIC_BASE_URL", "")
		t.Setenv("API_BASE_URL", "")
		t.Setenv("APP_BASE_URL", "")
		want := "https://api.locazar.com/uploads/products/x.jpg"
		if got := resolveFollowerImageURL(nil, "uploads/products/x.jpg"); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("S3 base is NOT used for uploads/ paths", func(t *testing.T) {
		t.Setenv("NOTIFICATION_PUBLIC_BASE_URL", "")
		t.Setenv("PUBLIC_BASE_URL", "")
		t.Setenv("API_BASE_URL", "")
		t.Setenv("APP_BASE_URL", "")
		t.Setenv("S3_PUBLIC_BASE_URL", "https://s3.example/bucket") // must be ignored
		want := "https://api.locazar.com/uploads/products/x.jpg"
		if got := resolveFollowerImageURL(nil, "uploads/products/x.jpg"); got != want {
			t.Fatalf("got %q, want %q (S3 base must not be used for uploads/)", got, want)
		}
	})

	t.Run("leading slash is normalised", func(t *testing.T) {
		t.Setenv("PUBLIC_BASE_URL", "https://api.locazar.com")
		want := "https://api.locazar.com/uploads/products/x.jpg"
		if got := resolveFollowerImageURL(nil, "/uploads/products/x.jpg"); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}

func TestFollowerFirstNonBlank(t *testing.T) {
	if got := followerFirstNonBlank([]string{"", "  ", "a", "b"}); got != "a" {
		t.Errorf("got %q, want \"a\"", got)
	}
	if got := followerFirstNonBlank(nil); got != "" {
		t.Errorf("got %q, want \"\"", got)
	}
}

// stubCloud stands in for the object storage service, resolving bare object
// keys against a CDN base the way objectStorage.PublicURL does.
type stubCloud struct{ base string }

func (s stubCloud) SaveFile(ctx context.Context, fh *multipart.FileHeader, opts cloud.SaveOptions) (string, error) {
	return "", nil
}
func (s stubCloud) SaveBytes(ctx context.Context, data []byte, opts cloud.SaveOptions) (string, error) {
	return "", nil
}
func (s stubCloud) PublicURL(objectKey string) string {
	if s.base == "" {
		return objectKey // mirrors noopObjectStorage: no storage configured
	}
	return s.base + "/" + strings.TrimLeft(objectKey, "/")
}
func (s stubCloud) PresignedURL(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	return "", nil
}
func (s stubCloud) DeleteObject(ctx context.Context, objectKey string) error { return nil }
func (s stubCloud) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	return nil, nil
}
func (s stubCloud) GetBytes(ctx context.Context, objectKey string) ([]byte, error) { return nil, nil }

// Seller product uploads store a bare object key ("products/x.jpg" from
// SaveBytes), not an "uploads/" path. Resolving those against the API host
// produced a URL that 404s, and FCM drops an image it cannot fetch — the
// notification arrived with no picture.
func TestResolveFollowerImageURL_ObjectKeys(t *testing.T) {
	t.Run("object key resolves against the CDN, not the API host", func(t *testing.T) {
		t.Setenv("PUBLIC_BASE_URL", "https://api.locazar.com")
		t.Setenv("S3_PUBLIC_BASE_URL", "")
		cs := stubCloud{base: "https://innoida.utho.io/locazar"}
		want := "https://innoida.utho.io/locazar/products/abc.jpg"
		if got := resolveFollowerImageURL(cs, "products/abc.jpg"); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("uploads path still uses the API host even with a cloud service", func(t *testing.T) {
		t.Setenv("NOTIFICATION_PUBLIC_BASE_URL", "")
		t.Setenv("PUBLIC_BASE_URL", "https://api.locazar.com")
		cs := stubCloud{base: "https://innoida.utho.io/locazar"}
		want := "https://api.locazar.com/uploads/products/x.jpg"
		if got := resolveFollowerImageURL(cs, "uploads/products/x.jpg"); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("no-op storage falls back to the S3 env base", func(t *testing.T) {
		t.Setenv("S3_PUBLIC_BASE_URL", "https://cdn.example/bucket")
		cs := stubCloud{} // PublicURL returns the key unchanged
		want := "https://cdn.example/bucket/products/abc.jpg"
		if got := resolveFollowerImageURL(cs, "products/abc.jpg"); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("unresolvable object key yields empty, never a wrong host", func(t *testing.T) {
		t.Setenv("PUBLIC_BASE_URL", "https://api.locazar.com")
		t.Setenv("S3_PUBLIC_BASE_URL", "")
		if got := resolveFollowerImageURL(nil, "products/abc.jpg"); got != "" {
			t.Fatalf("got %q, want \"\" (must not point at the API host)", got)
		}
	})

	t.Run("absolute url passes through regardless of storage", func(t *testing.T) {
		cs := stubCloud{base: "https://innoida.utho.io/locazar"}
		if got := resolveFollowerImageURL(cs, "https://cdn.x/a.jpg"); got != "https://cdn.x/a.jpg" {
			t.Fatalf("absolute url changed: %q", got)
		}
	})
}
