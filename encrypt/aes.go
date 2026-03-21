package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// aesEncryptCBC encrypts data using AES-256-CBC with PKCS#7 padding.
// A random 16-byte IV is prepended to the ciphertext.
func aesEncryptCBC(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("encrypt: aes cipher: %w", err)
	}

	// PKCS#7 padding
	padLen := aes.BlockSize - len(plaintext)%aes.BlockSize
	padded := make([]byte, len(plaintext)+padLen)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	// Random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("encrypt: random iv: %w", err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(padded))
	copy(ciphertext, iv)

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], padded)

	return ciphertext, nil
}

// aesEncryptECB encrypts exactly one block using AES-256-ECB (no padding, no IV).
// Used for the Perms entry.
func aesEncryptECB(key, plaintext []byte) ([]byte, error) {
	if len(plaintext) != aes.BlockSize {
		return nil, fmt.Errorf("encrypt: ECB input must be %d bytes, got %d", aes.BlockSize, len(plaintext))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("encrypt: aes cipher: %w", err)
	}
	ciphertext := make([]byte, aes.BlockSize)
	block.Encrypt(ciphertext, plaintext)
	return ciphertext, nil
}
