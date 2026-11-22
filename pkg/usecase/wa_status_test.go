package usecase

import (
	"encoding/json"
	"testing"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// A real-shaped Meta status webhook for a message that failed because the
// seller is outside the 24h window.
func TestDecodeFailedStatus(t *testing.T) {
	raw := `{"object":"whatsapp_business_account","entry":[{"changes":[{"field":"messages","value":{
		"messaging_product":"whatsapp",
		"statuses":[{"id":"wamid.OUT1","status":"failed","timestamp":"1","recipient_id":"919999999999",
			"errors":[{"code":131047,"title":"Re-engagement message","message":"Re-engagement message",
			"error_data":{"details":"Message failed to send because more than 24 hours have passed"}}]}]}}]}]}`
	var p domain.WAWebhookPayload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatal(err)
	}
	st := p.Entry[0].Changes[0].Value.Statuses
	if len(st) != 1 || st[0].ID != "wamid.OUT1" || st[0].Status != "failed" ||
		len(st[0].Errors) != 1 || st[0].Errors[0].Code != waErrOutsideWindow {
		t.Fatalf("decoded %+v", st)
	}
}
