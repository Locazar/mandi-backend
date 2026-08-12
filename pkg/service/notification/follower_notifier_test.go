package notification

import "testing"

func TestResolveFollowerImageURL(t *testing.T) {
	t.Run("absolute passes through", func(t *testing.T) {
		t.Setenv("NOTIFICATION_PUBLIC_BASE_URL", "https://ignored.example")
		if got := resolveFollowerImageURL("https://cdn.x/a.jpg"); got != "https://cdn.x/a.jpg" {
			t.Fatalf("absolute url changed: %q", got)
		}
	})

	t.Run("empty stays empty", func(t *testing.T) {
		if got := resolveFollowerImageURL(""); got != "" {
			t.Fatalf("expected empty, got %q", got)
		}
	})

	t.Run("relative joins the public base", func(t *testing.T) {
		t.Setenv("NOTIFICATION_PUBLIC_BASE_URL", "")
		t.Setenv("PUBLIC_BASE_URL", "https://api.locazar.com/")
		t.Setenv("S3_PUBLIC_BASE_URL", "")
		want := "https://api.locazar.com/uploads/products/x.jpg"
		if got := resolveFollowerImageURL("uploads/products/x.jpg"); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("relative with no env base falls back to the API host", func(t *testing.T) {
		t.Setenv("NOTIFICATION_PUBLIC_BASE_URL", "")
		t.Setenv("PUBLIC_BASE_URL", "")
		t.Setenv("API_BASE_URL", "")
		t.Setenv("APP_BASE_URL", "")
		want := "https://api.locazar.com/uploads/products/x.jpg"
		if got := resolveFollowerImageURL("uploads/products/x.jpg"); got != want {
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
		if got := resolveFollowerImageURL("uploads/products/x.jpg"); got != want {
			t.Fatalf("got %q, want %q (S3 base must not be used for uploads/)", got, want)
		}
	})

	t.Run("leading slash is normalised", func(t *testing.T) {
		t.Setenv("PUBLIC_BASE_URL", "https://api.locazar.com")
		want := "https://api.locazar.com/uploads/products/x.jpg"
		if got := resolveFollowerImageURL("/uploads/products/x.jpg"); got != want {
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
