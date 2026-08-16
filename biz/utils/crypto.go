package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

// LegacyDefaultKey 是 ENCRYPTION_KEY 机制引入前使用的硬编码开发密钥。
// 仅保留用于存量密文迁移（MigrateLegacyCiphertext），不得用于加密新数据。
const LegacyDefaultKey = "12345678901234567890123456789012"

var encryptionKey []byte

func InitEncryption() {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		// Fallback for dev, but in prod this should be set
		key = LegacyDefaultKey
	}
	k, err := NormalizeKey(key)
	if err != nil {
		// 密钥非法时必须快速失败，否则所有加解密都会静默出错
		panic(fmt.Sprintf("invalid ENCRYPTION_KEY: %v", err))
	}
	encryptionKey = k
}

// NormalizeKey 接受 16/24/32 字节原始密钥，或 64 位 hex 字符串（解码为
// 32 字节 AES-256 密钥，例如 `openssl rand -hex 32` 的输出）。
func NormalizeKey(key string) ([]byte, error) {
	if len(key) == 64 {
		if b, err := hex.DecodeString(key); err == nil {
			return b, nil
		}
	}
	switch len(key) {
	case 16, 24, 32:
		return []byte(key), nil
	}
	return nil, fmt.Errorf("key length %d is not 16/24/32 bytes or 64 hex chars", len(key))
}

// UsingLegacyKey 报告当前生效密钥是否仍为旧默认密钥（未配置 ENCRYPTION_KEY
// 的开发环境）。此时无需迁移存量密文。
func UsingLegacyKey() bool {
	return string(encryptionKey) == LegacyDefaultKey
}

// MigrateLegacyCiphertext 尝试用旧默认密钥解密并以当前密钥重新加密。
// 仅当旧密钥解出可打印 ASCII（token/密码均为 ASCII 文本）时才认定为存量
// 密文；否则原样返回并标记未迁移。CFB 模式无法从解密错误区分密钥，必须
// 依靠明文形态判断。
func MigrateLegacyCiphertext(cipherText string) (string, bool, error) {
	plain, err := decryptWith([]byte(LegacyDefaultKey), cipherText)
	if err != nil || !isPrintableASCII(plain) {
		return cipherText, false, nil
	}
	reEncrypted, err := encryptWith(encryptionKey, plain)
	if err != nil {
		return cipherText, false, err
	}
	return reEncrypted, true, nil
}

func isPrintableASCII(s string) bool {
	if s == "" {
		return false
	}
	for _, b := range []byte(s) {
		// 允许 \t \n \r（SSH 私钥等多行 PEM 文本）与可见 ASCII
		if b == '\t' || b == '\n' || b == '\r' {
			continue
		}
		if b < 0x20 || b > 0x7E {
			return false
		}
	}
	return true
}

func Encrypt(text string) (string, error) {
	if text == "" {
		return "", nil
	}
	return encryptWith(encryptionKey, text)
}

func Decrypt(cryptoText string) (string, error) {
	if cryptoText == "" {
		return "", nil
	}
	return decryptWith(encryptionKey, cryptoText)
}

func encryptWith(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
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

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func decryptWith(key []byte, cryptoText string) (string, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
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
