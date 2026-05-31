package crypto

import (
	"crypto/rand"
	"strings"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return k
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	svc, err := NewService(map[string][]byte{"v1": testKey(t)}, "v1")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	return svc
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	svc := newTestService(t)
	plain := "ACME1234567890"

	ct, err := svc.Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if ct == plain || !strings.HasPrefix(ct, "v1:") {
		t.Fatalf("ciphertext not key-tagged: %q", ct)
	}

	got, err := svc.Decrypt(ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != plain {
		t.Fatalf("round-trip = %q, want %q", got, plain)
	}
}

func TestEncryptNonDeterministic(t *testing.T) {
	svc := newTestService(t)
	a, _ := svc.Encrypt("same")
	b, _ := svc.Encrypt("same")
	if a == b {
		t.Fatal("expected distinct ciphertexts (random nonce)")
	}
}

func TestDecryptTamperDetected(t *testing.T) {
	svc := newTestService(t)
	ct, _ := svc.Encrypt("secret")
	// flip the last character of the base64 body
	tampered := ct[:len(ct)-1] + string(ct[len(ct)-1]^1)
	if _, err := svc.Decrypt(tampered); err == nil {
		t.Fatal("expected GCM authentication failure on tampered ciphertext")
	}
}

func TestKeyRotationDecryptsOldKey(t *testing.T) {
	k1, k2 := testKey(t), testKey(t)
	old, err := NewService(map[string][]byte{"v1": k1}, "v1")
	if err != nil {
		t.Fatal(err)
	}
	ct, _ := old.Encrypt("legacy")

	// New service has both keys; active is v2, but must still decrypt v1 data.
	rotated, err := NewService(map[string][]byte{"v1": k1, "v2": k2}, "v2")
	if err != nil {
		t.Fatal(err)
	}
	got, err := rotated.Decrypt(ct)
	if err != nil {
		t.Fatalf("Decrypt old ciphertext after rotation: %v", err)
	}
	if got != "legacy" {
		t.Fatalf("got %q, want legacy", got)
	}
	// New writes use the active key.
	newCt, _ := rotated.Encrypt("fresh")
	if !strings.HasPrefix(newCt, "v2:") {
		t.Fatalf("new ciphertext should use active key v2: %q", newCt)
	}
}

func TestNewServiceValidation(t *testing.T) {
	if _, err := NewService(map[string][]byte{}, "v1"); err == nil {
		t.Fatal("expected error: no keys")
	}
	if _, err := NewService(map[string][]byte{"v1": make([]byte, 16)}, "v1"); err == nil {
		t.Fatal("expected error: key not 32 bytes")
	}
	if _, err := NewService(map[string][]byte{"v1": testKey(t)}, "v2"); err == nil {
		t.Fatal("expected error: active key not in keyring")
	}
}

func TestParseKeyring(t *testing.T) {
	k := testKey(t)
	enc := "v1:" + b64(k)
	keys, err := ParseKeyring(enc)
	if err != nil {
		t.Fatalf("ParseKeyring: %v", err)
	}
	if len(keys) != 1 || len(keys["v1"]) != 32 {
		t.Fatalf("unexpected parse result: %+v", keys)
	}
}
