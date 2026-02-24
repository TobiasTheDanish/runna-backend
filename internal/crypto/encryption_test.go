package crypto

import (
	"strings"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// 32-byte key (256 bits)
	key := "12345678901234567890123456789012"
	plaintext := "strava_access_token_12345"

	// Encrypt
	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Verify it's different from plaintext
	if encrypted == plaintext {
		t.Fatal("Encrypted text should be different from plaintext")
	}

	// Decrypt
	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Verify we got back the original
	if decrypted != plaintext {
		t.Fatalf("Expected %s, got %s", plaintext, decrypted)
	}
}

func TestEncryptWithInvalidKeyLength(t *testing.T) {
	shortKey := "tooshort"
	plaintext := "test"

	_, err := Encrypt(plaintext, shortKey)
	if err == nil {
		t.Fatal("Expected error for short key, got nil")
	}

	if !strings.Contains(err.Error(), "must be 32 bytes") {
		t.Fatalf("Expected key length error, got: %v", err)
	}
}

func TestDecryptWithInvalidKeyLength(t *testing.T) {
	shortKey := "tooshort"
	ciphertext := "test"

	_, err := Decrypt(ciphertext, shortKey)
	if err == nil {
		t.Fatal("Expected error for short key, got nil")
	}

	if !strings.Contains(err.Error(), "must be 32 bytes") {
		t.Fatalf("Expected key length error, got: %v", err)
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := "12345678901234567890123456789012"
	key2 := "abcdefghijklmnopqrstuvwxyz123456"
	plaintext := "secret_token"

	// Encrypt with key1
	encrypted, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Try to decrypt with key2
	_, err = Decrypt(encrypted, key2)
	if err == nil {
		t.Fatal("Expected error when decrypting with wrong key, got nil")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	key := "12345678901234567890123456789012"
	invalidCiphertext := "not-valid-base64!"

	_, err := Decrypt(invalidCiphertext, key)
	if err == nil {
		t.Fatal("Expected error for invalid ciphertext, got nil")
	}
}

func TestEncryptDifferentOutputs(t *testing.T) {
	key := "12345678901234567890123456789012"
	plaintext := "same_text"

	// Encrypt twice
	encrypted1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	encrypted2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	// Should produce different ciphertexts (due to random nonce)
	if encrypted1 == encrypted2 {
		t.Fatal("Multiple encryptions of same plaintext should produce different ciphertexts")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := Decrypt(encrypted1, key)
	if err != nil {
		t.Fatalf("First decryption failed: %v", err)
	}

	decrypted2, err := Decrypt(encrypted2, key)
	if err != nil {
		t.Fatalf("Second decryption failed: %v", err)
	}

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Fatal("Both decryptions should produce original plaintext")
	}
}

func TestEncryptEmptyString(t *testing.T) {
	key := "12345678901234567890123456789012"
	plaintext := ""

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption of empty string failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("Expected empty string, got %s", decrypted)
	}
}

func TestEncryptLongString(t *testing.T) {
	key := "12345678901234567890123456789012"
	// Create a long token string (typical OAuth token)
	plaintext := strings.Repeat("a", 500)

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption of long string failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatal("Decrypted text doesn't match original long string")
	}
}

func BenchmarkEncrypt(b *testing.B) {
	key := "12345678901234567890123456789012"
	plaintext := "strava_access_token_12345678901234567890"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encrypt(plaintext, key)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key := "12345678901234567890123456789012"
	plaintext := "strava_access_token_12345678901234567890"

	encrypted, _ := Encrypt(plaintext, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decrypt(encrypted, key)
	}
}
