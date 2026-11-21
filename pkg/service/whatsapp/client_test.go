package whatsapp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, status int, body string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	c := NewClient("123", "token")
	c.http = srv.Client()
	c.http.Transport = rewrite{srv.URL}
	return c
}

type rewrite struct{ base string }

func (r rewrite) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	u := *req.URL
	u.Scheme = "http"
	u.Host = strings.TrimPrefix(r.base, "http://")
	req2.URL = &u
	return http.DefaultTransport.RoundTrip(req2)
}

func TestSendTextWithIDReturnsMessageID(t *testing.T) {
	c := newTestClient(t, 200, `{"messages":[{"id":"wamid.ABC"}]}`)
	id, err := c.SendTextWithID("919876543210", "hi")
	if err != nil || id != "wamid.ABC" {
		t.Fatalf("got id=%q err=%v", id, err)
	}
}

func TestOutsideWindowError(t *testing.T) {
	c := newTestClient(t, 400, `{"error":{"code":131047,"message":"Re-engagement message"}}`)
	_, err := c.SendTextWithID("919876543210", "hi")
	if !IsOutsideWindow(err) {
		t.Fatalf("expected outside-window error, got %v", err)
	}
}

func TestTemplateSafe(t *testing.T) {
	if got := templateSafe("line1\nline2\t      x"); got != "line1 line2 x" {
		t.Fatalf("templateSafe = %q", got)
	}
	if templateSafe("  ") != "-" {
		t.Fatal("empty param should become -")
	}
}
