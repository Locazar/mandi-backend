// Package crypto provides application-level envelope encryption for PII at rest
// (AES-256-GCM) with a versioned keyring so keys can be rotated without bulk
// re-encryption. Ciphertext is self-describing: "<key_id>:<base64(nonce||ct)>".
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const keyIDSeparator = ":"

// Service encrypts and decrypts PII using a keyring keyed by id.
type Service struct {
	keys      map[string][]byte
	activeKey string
}

// NewService builds a Service from a keyring. Every key must be 32 bytes
// (AES-256) and the active key id must exist in the keyring.
func NewService(keys map[string][]byte, activeKey string) (*Service, error) {
	if len(keys) == 0 {
		return nil, errors.New("crypto: keyring is empty")
	}
	for id, k := range keys {
		if len(k) != 32 {
			return nil, fmt.Errorf("crypto: key %q must be 32 bytes, got %d", id, len(k))
		}
		if strings.Contains(id, keyIDSeparator) {
			return nil, fmt.Errorf("crypto: key id %q must not contain %q", id, keyIDSeparator)
		}
	}
	if _, ok := keys[activeKey]; !ok {
		return nil, fmt.Errorf("crypto: active key %q not present in keyring", activeKey)
	}
	return &Service{keys: keys, activeKey: activeKey}, nil
}

// Encrypt seals plaintext with the active key, returning "<key_id>:<base64>".
func (s *Service) Encrypt(plaintext string) (string, error) {
	gcm, err := s.gcm(s.activeKey)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return s.activeKey + keyIDSeparator + b64(sealed), nil
}

// Decrypt opens a "<key_id>:<base64>" ciphertext using the key it names.
func (s *Service) Decrypt(ciphertext string) (string, error) {
	id, body, ok := strings.Cut(ciphertext, keyIDSeparator)
	if !ok {
		return "", errors.New("crypto: malformed ciphertext (missing key id)")
	}
	gcm, err := s.gcm(id)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return "", fmt.Errorf("crypto: invalid base64: %w", err)
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("crypto: ciphertext too short")
	}
	nonce, sealed := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: decryption failed: %w", err)
	}
	return string(plain), nil
}

func (s *Service) gcm(keyID string) (cipher.AEAD, error) {
	key, ok := s.keys[keyID]
	if !ok {
		return nil, fmt.Errorf("crypto: unknown key id %q", keyID)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// ParseKeyring decodes "id1:base64key1,id2:base64key2" into a keyring.
func ParseKeyring(spec string) (map[string][]byte, error) {
	keys := make(map[string][]byte)
	for _, entry := range strings.Split(spec, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		id, encoded, ok := strings.Cut(entry, keyIDSeparator)
		if !ok {
			return nil, fmt.Errorf("crypto: malformed keyring entry %q", entry)
		}
		key, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("crypto: key %q invalid base64: %w", id, err)
		}
		keys[id] = key
	}
	if len(keys) == 0 {
		return nil, errors.New("crypto: no keys parsed from keyring spec")
	}
	return keys, nil
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
