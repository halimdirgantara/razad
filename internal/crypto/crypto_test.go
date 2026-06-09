package crypto

import (
	"testing"
)

func TestNew_EmptySecret(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Error("expected error for empty secret")
	}
}

func TestNewFromBytes_InvalidKeyLength(t *testing.T) {
	_, err := NewFromBytes([]byte("short"))
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	e, err := New("test-secret-key-12345")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	plaintext := []byte("sensitive-data-here")
	cipherHex, err := e.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if cipherHex == "" {
		t.Fatal("expected non-empty ciphertext")
	}

	decrypted, err := e.Decrypt(cipherHex)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("round-trip mismatch: got %s, want %s", decrypted, plaintext)
	}
}

func TestEncrypt_DifferentCipherEachTime(t *testing.T) {
	e, err := New("test-secret-key-12345")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	plaintext := []byte("same-data")
	c1, _ := e.Encrypt(plaintext)
	c2, _ := e.Encrypt(plaintext)

	if c1 == c2 {
		t.Error("expected different ciphertexts due to random nonce")
	}
}

func TestDecrypt_InvalidCiphertext(t *testing.T) {
	e, err := New("test-secret-key-12345")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_, err = e.Decrypt("invalid-hex")
	if err == nil {
		t.Error("expected error for invalid hex ciphertext")
	}

	_, err = e.Decrypt("abcd")
	if err == nil {
		t.Error("expected error for too-short ciphertext")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	e1, _ := New("correct-key-1234567890")
	e2, _ := New("wrong-key-1234567890")

	cipherHex, err := e1.Encrypt([]byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = e2.Decrypt(cipherHex)
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}
}

func TestMaskString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc", "***"},
		{"abcd", "****"},
		{"abcde", "a***e"},
		{"abcdefgh", "a******h"},
		{"", ""},
		{"a", "a"},
		{"ab", "ab"},
	}

	for _, tt := range tests {
		got := MaskString(tt.input)
		if got != tt.want {
			t.Errorf("MaskString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
