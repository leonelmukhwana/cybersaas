package mpesa

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

type Encryptor struct {
	key []byte
}

func NewEncryptor() (*Encryptor, error) {
	value := os.Getenv("MPESA_CONFIG_ENCRYPTION_KEY")

	if value == "" {
		return nil, errors.New(
			"MPESA_CONFIG_ENCRYPTION_KEY is required",
		)
	}

	key, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf(
			"decode MPESA_CONFIG_ENCRYPTION_KEY: %w",
			err,
		)
	}

	if len(key) != 32 {
		return nil, errors.New(
			"MPESA_CONFIG_ENCRYPTION_KEY must decode to exactly 32 bytes",
		)
	}

	return &Encryptor{key: key}, nil
}

func (e *Encryptor) Encrypt(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create AES-GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate encryption nonce: %w", err)
	}

	ciphertext := gcm.Seal(
		nonce,
		nonce,
		[]byte(value),
		nil,
	)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryptor) Decrypt(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	raw, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf(
			"decode encrypted value: %w",
			err,
		)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create AES-GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()

	if len(raw) < nonceSize {
		return "", errors.New("encrypted value is invalid")
	}

	nonce := raw[:nonceSize]
	ciphertext := raw[nonceSize:]

	plaintext, err := gcm.Open(
		nil,
		nonce,
		ciphertext,
		nil,
	)
	if err != nil {
		return "", errors.New("decrypt value failed")
	}

	return string(plaintext), nil
}
