package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

type AESEncryptor struct {
	key []byte
}

func NewAESEncryptor(key string) (*AESEncryptor, error) {
	if len(key) < 32 {
		return nil, errors.New("encryption key must be at least 32 characters")
	}
	keyBytes := []byte(key)[:32]
	return &AESEncryptor{key: keyBytes}, nil
}

func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

var defaultEncryptor *AESEncryptor

func Init(key string) error {
	enc, err := NewAESEncryptor(key)
	if err != nil {
		return err
	}
	defaultEncryptor = enc
	return nil
}

func Encrypt(plaintext string) (string, error) {
	if defaultEncryptor == nil {
		return "", errors.New("encryptor not initialized")
	}
	return defaultEncryptor.Encrypt(plaintext)
}

func Decrypt(ciphertext string) (string, error) {
	if defaultEncryptor == nil {
		return "", errors.New("encryptor not initialized")
	}
	return defaultEncryptor.Decrypt(ciphertext)
}
