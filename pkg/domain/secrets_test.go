package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSecretsNotSerialized(t *testing.T) {
	u, _ := json.Marshal(User{Password: "hunter2"})
	a, _ := json.Marshal(Admin{Password: "hunter2"})
	r, _ := json.Marshal(AdminRefreshSession{RefreshToken: "tok"})
	for _, b := range [][]byte{u, a, r} {
		if strings.Contains(string(b), "hunter2") || strings.Contains(string(b), "\"tok\"") {
			t.Fatalf("secret leaked in JSON: %s", b)
		}
		if strings.Contains(string(b), "password") || strings.Contains(string(b), "refresh_token") {
			t.Fatalf("secret field name still serialized: %s", b)
		}
	}
}
