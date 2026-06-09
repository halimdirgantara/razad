// Package crypto provides encryption at rest for secrets and sensitive values.
// Uses AES-256-GCM with key derivation for secure storage.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Common errors.
var (
	ErrInvalidKey     = errors.New("crypto: invalid key")
	ErrInvalidCipher  = errors.New("crypto: invalid ciphertext")
	ErrDecryptionFail = errors.New("crypto: decryption failed")
)

// Encrypter provides encryption and decryption operations.
type Encrypter struct {
	key []byte
}

// New creates an Encrypter from a raw secret key. The key is derived through
// SHA-256 to produce a consistent 32-byte AES-256 key.
func New(secret string) (*Encrypter, error) {
	if secret == "" {
		return nil, fmt.Errorf("%w: secret must not be empty", ErrInvalidKey)
	}
	// Derive a 32-byte key using SHA-256.
	hash := sha256.Sum256([]byte(secret))
	return &Encrypter{key: hash[:]}, nil
}

// NewFromBytes creates an Encrypter from a 32-byte key.
func NewFromBytes(key []byte) (*Encrypter, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("%w: key must be 32 bytes, got %d", ErrInvalidKey, len(key))
	}
	k := make([]byte, 32)
	copy(k, key)
	return &Encrypter{key: k}, nil
}

// Encrypt encrypts plaintext and returns a hex-encoded ciphertext string.
// Format: nonce (12 bytes) + ciphertext.
func (e *Encrypter) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("crypto: failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a hex-encoded ciphertext produced by Encrypt.
func (e *Encrypter) Decrypt(cipherHex string) ([]byte, error) {
	ciphertext, err := hex.DecodeString(cipherHex)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCipher, err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("crypto: failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidCipher
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("crypto: %w: %v", ErrDecryptionFail, err)
	}

	return plaintext, nil
}

// MaskString returns a masked version of a string for safe logging.
// Strings of length 0-2 are shown as-is. Strings of length 3-4 are fully masked.
// Strings of length 5+ show only the first and last character with '*' between.
func MaskString(s string) string {
	n := len(s)
	switch {
	case n <= 2:
		return s
	case n <= 4:
		return strings.Repeat("*", n)
	default:
		return s[:1] + strings.Repeat("*", n-2) + s[n-1:]
	}
}
