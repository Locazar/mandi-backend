package handler

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// A shop_socials row exists as soon as a customer follows or likes a shop, with
// no rating/review/comments on it. Such a row must NOT count as feedback: the
// customer apps treat a non-empty feedback response as "already reviewed" and
// hide the review form, so counting a follow would lock a customer out of ever
// reviewing a shop they follow.
func TestShopSocialHasFeedback(t *testing.T) {
	tests := []struct {
		name string
		row  domain.ShopSocial
		want bool
	}{
		{"follow only is not feedback", domain.ShopSocial{IsFollower: true}, false},
		{"like only is not feedback", domain.ShopSocial{IsLiked: true}, false},
		{"empty row is not feedback", domain.ShopSocial{}, false},
		{"whitespace review is not feedback", domain.ShopSocial{Review: "   \n\t "}, false},
		{"whitespace comments is not feedback", domain.ShopSocial{Comments: "  "}, false},
		{"rating is feedback", domain.ShopSocial{Rating: 4}, true},
		{"review is feedback", domain.ShopSocial{Review: "great shop"}, true},
		{"comments is feedback", domain.ShopSocial{Comments: "fast delivery"}, true},
		{"follower who also rated is feedback", domain.ShopSocial{IsFollower: true, Rating: 5}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shopSocialHasFeedback(tt.row); got != tt.want {
				t.Fatalf("shopSocialHasFeedback(%+v) = %v, want %v", tt.row, got, tt.want)
			}
		})
	}
}

// The customer apps decode this payload with a `customer_id` key, while the
// table column is user_id — the mapping is what makes the response decodable.
// admin_id and the follow/like flags are deliberately not carried over.
func TestToShopFeedbackMapsUserIDToCustomerID(t *testing.T) {
	now := time.Now()
	row := domain.ShopSocial{
		ID:         "row1",
		ShopID:     "shop1",
		AdminID:    "admin1",
		UserID:     "user1",
		IsFollower: true,
		IsLiked:    true,
		Rating:     5,
		Review:     "  excellent  ",
		Comments:   "  quick  ",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	got := toShopFeedback(row)

	if got.CustomerID != "user1" {
		t.Errorf("CustomerID = %q, want %q", got.CustomerID, "user1")
	}
	if got.ShopID != "shop1" {
		t.Errorf("ShopID = %q, want %q", got.ShopID, "shop1")
	}
	if got.Rating != 5 {
		t.Errorf("Rating = %d, want 5", got.Rating)
	}
	// Trimmed so padded input does not reach the apps as-is.
	if got.Review != "excellent" {
		t.Errorf("Review = %q, want %q", got.Review, "excellent")
	}
	if got.Comments != "quick" {
		t.Errorf("Comments = %q, want %q", got.Comments, "quick")
	}
	if !got.CreatedAt.Equal(now) || !got.UpdatedAt.Equal(now) {
		t.Errorf("timestamps not carried over: %+v", got)
	}
}

// Contract guard for GET /social/feedback/shop/{shop_id}. The customer apps
// decode this as ApiEnvelope<[Feedback]> — `data` MUST be a JSON array whose
// elements carry snake_case keys matching Feedback's fields. This previously
// returned an object ({shop_id, feedback}), which no client could decode; a
// regression to any wrapped shape would silently break the review flow again.
func TestShopFeedbackResponseShapeMatchesClientContract(t *testing.T) {
	body, err := json.Marshal(response.Response{
		Status:  true,
		Message: "Successfully got shop feedback",
		Data:    []response.ShopFeedback{toShopFeedback(domain.ShopSocial{ShopID: "s1", UserID: "u1", Rating: 4, Review: "good"})},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var envelope struct {
		Success bool              `json:"success"`
		Data    []json.RawMessage `json:"data"` // fails to decode if data is not an array
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("data is not a JSON array, clients cannot decode it: %v (body=%s)", err, body)
	}
	if len(envelope.Data) != 1 {
		t.Fatalf("data length = %d, want 1 (body=%s)", len(envelope.Data), body)
	}

	var item map[string]any
	if err := json.Unmarshal(envelope.Data[0], &item); err != nil {
		t.Fatalf("unmarshal item: %v", err)
	}
	// Non-optional on the client side: absence breaks the whole decode.
	for _, key := range []string{"shop_id", "customer_id", "rating"} {
		if _, ok := item[key]; !ok {
			t.Errorf("missing required key %q in %v", key, item)
		}
	}
	// Internal fields must not leak into a customer-facing payload.
	for _, key := range []string{"admin_id", "user_id", "is_follower", "is_liked"} {
		if _, ok := item[key]; ok {
			t.Errorf("internal key %q leaked into feedback payload: %v", key, item)
		}
	}
}

// No feedback must serialize as [] rather than null: the apps decode a
// non-optional array and read `isEmpty` to decide whether to show the form.
func TestEmptyShopFeedbackSerializesAsArrayNotNull(t *testing.T) {
	body, err := json.Marshal(response.Response{Data: []response.ShopFeedback{}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(body), `"data":[]`) {
		t.Fatalf("empty feedback should serialize as [], got %s", body)
	}
}
