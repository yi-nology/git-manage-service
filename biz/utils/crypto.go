package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"os"

	"git.enjoye.top/enjoydream/ekit/pkg/encoding"
)

var encryptionKey []byte

func InitEncryption() {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		// Fallback for dev, but in prod this should be set
		key = "12345678901234567890123456789012" // 32 bytes for AES-256
	}
	encryptionKey = []byte(key)
}

func Encrypt(text string) (string, error) {
	if text == "" {
		return "", nil
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	b := []byte(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], b)

	return encoding.Base64URLEncode(string(ciphertext)), nil
}

func Decrypt(cryptoText string) (string, error) {
	if cryptoText == "" {
		return "", nil
	}
	ciphertextStr, err := encoding.Base64URLDecode(cryptoText)
	if err != nil {
		return "", err
	}
	ciphertext := []byte(ciphertextStr)

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}
