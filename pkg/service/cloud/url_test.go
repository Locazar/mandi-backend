package cloud

import (
	"context"
	"mime/multipart"
	"testing"
	"time"
)

type stubCS struct{ base string }

func (s stubCS) SaveFile(ctx context.Context, fh *multipart.FileHeader, opts SaveOptions) (string, error) {
	return "", nil
}
func (s stubCS) SaveBytes(ctx context.Context, data []byte, opts SaveOptions) (string, error) {
	return "", nil
}
func (s stubCS) PublicURL(k string) string { return s.base + "/" + k }
func (s stubCS) PresignedURL(ctx context.Context, k string, ttl time.Duration) (string, error) {
	return "", nil
}
func (s stubCS) DeleteObject(ctx context.Context, k string) error { return nil }

func TestResolveURL(t *testing.T) {
	cs := stubCS{base: "https://innoida.utho.io/locazar-dev"}
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"https://x.com/y.jpg", "https://x.com/y.jpg"},
		{"http://x.com/y.jpg", "http://x.com/y.jpg"},
		{"uploads/products/foo.jpg", "/uploads/products/foo.jpg"},
		{"/uploads/products/foo.jpg", "/uploads/products/foo.jpg"},
		{"products/bar.jpg", "https://innoida.utho.io/locazar-dev/products/bar.jpg"},
		{"admin-profiles/abc.jpg", "https://innoida.utho.io/locazar-dev/admin-profiles/abc.jpg"},
		{"offers/thumbnail/x.png", "https://innoida.utho.io/locazar-dev/offers/thumbnail/x.png"},
	}
	for _, c := range cases {
		if got := ResolveURL(cs, c.in); got != c.want {
			t.Errorf("ResolveURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
