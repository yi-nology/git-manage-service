package utils

import (
	"encoding/base64"
	"testing"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	InitEncryption()
	original := "hello world 1234!@#$"
	encrypted, err := Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if encrypted == "" {
		t.Fatal("Encrypt returned empty string for non-empty input")
	}
	if encrypted == original {
		t.Error("Encrypted text should differ from original")
	}
	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decrypted != original {
		t.Errorf("Round-trip mismatch: got %q, want %q", decrypted, original)
	}
}

func TestEncrypt_EmptyString(t *testing.T) {
	InitEncryption()
	result, err := Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt empty string should not error: %v", err)
	}
	if result != "" {
		t.Errorf("Encrypt empty string should return empty, got %q", result)
	}
}

func TestDecrypt_EmptyString(t *testing.T) {
	InitEncryption()
	result, err := Decrypt("")
	if err != nil {
		t.Fatalf("Decrypt empty string should not error: %v", err)
	}
	if result != "" {
		t.Errorf("Decrypt empty string should return empty, got %q", result)
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	InitEncryption()
	_, err := Decrypt("!!!invalid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64 input")
	}
}

func TestDecrypt_CiphertextTooShort(t *testing.T) {
	InitEncryption()
	short := base64.URLEncoding.EncodeToString([]byte("ab"))
	_, err := Decrypt(short)
	if err == nil {
		t.Error("expected error for too-short ciphertext")
	}
}

func TestEncryptDecrypt_LongText(t *testing.T) {
	InitEncryption()
	longStr := ""
	for i := 0; i < 1000; i++ {
		longStr += "This is a longer text string for testing. "
	}
	encrypted, err := Encrypt(longStr)
	if err != nil {
		t.Fatalf("Encrypt long text failed: %v", err)
	}
	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt long text failed: %v", err)
	}
	if decrypted != longStr {
		t.Error("Round-trip failed for long text")
	}
}

func TestEncryptDecrypt_SpecialCharacters(t *testing.T) {
	InitEncryption()
	specials := []string{
		"中文测试",
		"emoji 🎉🚀",
		"new\nlines\tand\ttabs",
		"{\"json\": true, \"key\": \"value\"}",
		"<html>&amp;entities</html>",
	}
	for _, s := range specials {
		enc, err := Encrypt(s)
		if err != nil {
			t.Fatalf("Encrypt %q failed: %v", s, err)
		}
		dec, err := Decrypt(enc)
		if err != nil {
			t.Fatalf("Decrypt %q failed: %v", s, err)
		}
		if dec != s {
			t.Errorf("Round-trip mismatch for %q: got %q", s, dec)
		}
	}
}

func TestEncrypt_DifferentCiphertexts(t *testing.T) {
	InitEncryption()
	same := "same plaintext"
	enc1, _ := Encrypt(same)
	enc2, _ := Encrypt(same)
	if enc1 == enc2 {
		t.Error("Encrypting same plaintext twice should produce different ciphertexts (random IV)")
	}
	dec1, _ := Decrypt(enc1)
	dec2, _ := Decrypt(enc2)
	if dec1 != dec2 {
		t.Error("Both ciphertexts should decrypt to the same plaintext")
	}
}
