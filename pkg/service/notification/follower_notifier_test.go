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

	t.Run("relative with no base is dropped (text-only push)", func(t *testing.T) {
		t.Setenv("NOTIFICATION_PUBLIC_BASE_URL", "")
		t.Setenv("PUBLIC_BASE_URL", "")
		t.Setenv("S3_PUBLIC_BASE_URL", "")
		if got := resolveFollowerImageURL("uploads/products/x.jpg"); got != "" {
			t.Fatalf("expected empty (image dropped), got %q", got)
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
