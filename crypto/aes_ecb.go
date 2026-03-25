package crypto

import (
	"bytes"
	"crypto/aes"
	"fmt"
)

// EncryptAESECB encrypts data using AES-128-ECB with PKCS7 padding
func EncryptAESECB(plaintext, key []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("key must be 16 bytes for AES-128")
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Add PKCS7 padding
	padded := pkcs7Pad(plaintext, aes.BlockSize)

	// Encrypt each block
	ciphertext := make([]byte, len(padded))
	for i := 0; i < len(padded); i += aes.BlockSize {
		block.Encrypt(ciphertext[i:i+aes.BlockSize], padded[i:i+aes.BlockSize])
	}

	return ciphertext, nil
}

// DecryptAESECB decrypts data using AES-128-ECB with PKCS7 padding
func DecryptAESECB(ciphertext, key []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("key must be 16 bytes for AES-128")
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext length must be multiple of block size")
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt each block
	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += aes.BlockSize {
		block.Decrypt(plaintext[i:i+aes.BlockSize], ciphertext[i:i+aes.BlockSize])
	}

	// Remove PKCS7 padding
	unpadded, err := pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad: %w", err)
	}

	return unpadded, nil
}

// AESECBPaddedSize computes the ciphertext size with PKCS7 padding
func AESECBPaddedSize(plaintextSize int) int {
	// PKCS7 always adds at least 1 byte of padding
	// Round up to next 16-byte boundary
	return ((plaintextSize + aes.BlockSize) / aes.BlockSize) * aes.BlockSize
}

// pkcs7Pad adds PKCS7 padding to data
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpad removes PKCS7 padding from data
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	if len(data)%blockSize != 0 {
		return nil, fmt.Errorf("data length must be multiple of block size")
	}

	padding := int(data[len(data)-1])

	if padding == 0 || padding > blockSize {
		return nil, fmt.Errorf("invalid padding value: %d", padding)
	}

	if padding > len(data) {
		return nil, fmt.Errorf("padding size exceeds data length")
	}

	// Verify padding
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}

	return data[:len(data)-padding], nil
}
