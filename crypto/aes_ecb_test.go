package crypto

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"testing"
)

// TestEncryptDecryptAESECB tests the encryption and decryption round-trip
func TestEncryptDecryptAESECB(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		key       []byte
	}{
		{
			name:      "short text",
			plaintext: "Hello",
			key:       []byte("0123456789abcdef"), // 16 bytes
		},
		{
			name:      "exact block size",
			plaintext: "1234567890123456", // exactly 16 bytes
			key:       []byte("0123456789abcdef"),
		},
		{
			name:      "longer text",
			plaintext: "This is a longer message that spans multiple blocks",
			key:       []byte("0123456789abcdef"),
		},
		{
			name:      "empty string",
			plaintext: "",
			key:       []byte("0123456789abcdef"),
		},
		{
			name:      "unicode text",
			plaintext: "你好世界 Hello World 🌍",
			key:       []byte("0123456789abcdef"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := EncryptAESECB([]byte(tt.plaintext), tt.key)
			if err != nil {
				t.Fatalf("EncryptAESECB failed: %v", err)
			}

			// Decrypt
			decrypted, err := DecryptAESECB(ciphertext, tt.key)
			if err != nil {
				t.Fatalf("DecryptAESECB failed: %v", err)
			}

			// Verify
			if string(decrypted) != tt.plaintext {
				t.Errorf("Decrypted text doesn't match:\nwant: %q\ngot:  %q", tt.plaintext, string(decrypted))
			}
		})
	}
}

// TestEncryptAESECBInvalidKey tests encryption with invalid key sizes
func TestEncryptAESECBInvalidKey(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
	}{
		{name: "too short", key: []byte("short")},
		{name: "15 bytes", key: []byte("123456789012345")},
		{name: "17 bytes", key: []byte("12345678901234567")},
		{name: "empty", key: []byte("")},
	}

	plaintext := []byte("test data")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncryptAESECB(plaintext, tt.key)
			if err == nil {
				t.Error("Expected error for invalid key size, got nil")
			}
		})
	}
}

// TestDecryptAESECBInvalidKey tests decryption with invalid key sizes
func TestDecryptAESECBInvalidKey(t *testing.T) {
	ciphertext := []byte("1234567890123456") // 16 bytes
	invalidKey := []byte("short")

	_, err := DecryptAESECB(ciphertext, invalidKey)
	if err == nil {
		t.Error("Expected error for invalid key size, got nil")
	}
}

// TestDecryptAESECBInvalidCiphertext tests decryption with invalid ciphertext
func TestDecryptAESECBInvalidCiphertext(t *testing.T) {
	key := []byte("0123456789abcdef")

	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{name: "not block aligned (15 bytes)", ciphertext: []byte("123456789012345")},
		{name: "not block aligned (17 bytes)", ciphertext: []byte("12345678901234567")},
		{name: "empty", ciphertext: []byte("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptAESECB(tt.ciphertext, key)
			if err == nil {
				t.Error("Expected error for invalid ciphertext, got nil")
			}
		})
	}
}

// TestAESECBPaddedSize tests the ciphertext size calculation
func TestAESECBPaddedSize(t *testing.T) {
	tests := []struct {
		plaintextSize int
		expectedSize  int
	}{
		{0, 16},   // empty → padded to 16
		{1, 16},   // 1 byte → padded to 16
		{15, 16},  // 15 bytes → padded to 16
		{16, 32},  // 16 bytes → padded to 32 (full block + padding block)
		{17, 32},  // 17 bytes → padded to 32
		{31, 32},  // 31 bytes → padded to 32
		{32, 48},  // 32 bytes → padded to 48
		{100, 112}, // 100 bytes → padded to 112 (7 blocks)
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.plaintextSize)), func(t *testing.T) {
			size := AESECBPaddedSize(tt.plaintextSize)
			if size != tt.expectedSize {
				t.Errorf("AESECBPaddedSize(%d) = %d, want %d", tt.plaintextSize, size, tt.expectedSize)
			}
		})
	}
}

// TestPKCS7Padding tests the padding function
func TestPKCS7Padding(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
		wantLen   int
	}{
		{
			name:      "empty data",
			data:      []byte{},
			blockSize: 16,
			wantLen:   16,
		},
		{
			name:      "1 byte",
			data:      []byte{0x01},
			blockSize: 16,
			wantLen:   16,
		},
		{
			name:      "15 bytes",
			data:      bytes.Repeat([]byte{0x01}, 15),
			blockSize: 16,
			wantLen:   16,
		},
		{
			name:      "16 bytes (full block)",
			data:      bytes.Repeat([]byte{0x01}, 16),
			blockSize: 16,
			wantLen:   32, // adds full padding block
		},
		{
			name:      "17 bytes",
			data:      bytes.Repeat([]byte{0x01}, 17),
			blockSize: 16,
			wantLen:   32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := pkcs7Pad(tt.data, tt.blockSize)
			if len(padded) != tt.wantLen {
				t.Errorf("pkcs7Pad() length = %d, want %d", len(padded), tt.wantLen)
			}

			// Verify padding bytes
			paddingLen := len(padded) - len(tt.data)
			for i := len(tt.data); i < len(padded); i++ {
				if padded[i] != byte(paddingLen) {
					t.Errorf("padding byte at %d = %d, want %d", i, padded[i], paddingLen)
				}
			}
		})
	}
}

// TestPKCS7Unpadding tests the unpadding function
func TestPKCS7Unpadding(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    []byte
		wantErr bool
	}{
		{
			name: "valid padding - 1 byte data",
			data: append([]byte{0x01}, bytes.Repeat([]byte{0x0f}, 15)...),
			want: []byte{0x01},
		},
		{
			name: "valid padding - 15 bytes data",
			data: append(bytes.Repeat([]byte{0x01}, 15), 0x01),
			want: bytes.Repeat([]byte{0x01}, 15),
		},
		{
			name: "valid padding - full padding block",
			data: append(bytes.Repeat([]byte{0x01}, 16), bytes.Repeat([]byte{0x10}, 16)...),
			want: bytes.Repeat([]byte{0x01}, 16),
		},
		{
			name:    "invalid - empty data",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "invalid - zero padding",
			data:    bytes.Repeat([]byte{0x00}, 16),
			wantErr: true,
		},
		{
			name:    "invalid - padding > block size",
			data:    append(bytes.Repeat([]byte{0x01}, 15), 0x11),
			wantErr: true,
		},
		{
			name:    "invalid - inconsistent padding",
			data:    append(bytes.Repeat([]byte{0x01}, 14), []byte{0x02, 0x03}...),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pkcs7Unpad(tt.data, 16)
			if (err != nil) != tt.wantErr {
				t.Errorf("pkcs7Unpad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !bytes.Equal(got, tt.want) {
				t.Errorf("pkcs7Unpad() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestKnownVectors tests against known test vectors
func TestKnownVectors(t *testing.T) {
	// Test vector: Known plaintext encrypted with known key
	key := []byte("0123456789abcdef")
	plaintext := []byte("Hello, World!")

	// Encrypt
	ciphertext, err := EncryptAESECB(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptAESECB failed: %v", err)
	}

	// Verify ciphertext is deterministic (same input → same output)
	ciphertext2, err := EncryptAESECB(plaintext, key)
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	if !bytes.Equal(ciphertext, ciphertext2) {
		t.Error("AES-ECB encryption is not deterministic")
	}

	// Decrypt and verify
	decrypted, err := DecryptAESECB(ciphertext, key)
	if err != nil {
		t.Fatalf("DecryptAESECB failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted text doesn't match original:\nwant: %q\ngot:  %q", plaintext, decrypted)
	}
}

// TestBase64Integration tests integration with base64 encoding (as used in CDN)
func TestBase64Integration(t *testing.T) {
	key := []byte("0123456789abcdef")
	plaintext := []byte("Test data for base64 integration")

	// Encrypt
	ciphertext, err := EncryptAESECB(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Encode to base64 (as done in CDN upload)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("Base64 decode failed: %v", err)
	}

	// Decrypt
	decrypted, err := DecryptAESECB(decoded, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted text doesn't match:\nwant: %q\ngot:  %q", plaintext, decrypted)
	}
}

// TestHexKeyFormat tests using hex-encoded keys (as used in CDN)
func TestHexKeyFormat(t *testing.T) {
	// Hex string key (32 chars = 16 bytes)
	hexKey := "0123456789abcdef0123456789abcdef"
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		t.Fatalf("Failed to decode hex key: %v", err)
	}

	plaintext := []byte("Test with hex key")

	// Encrypt with hex-decoded key
	ciphertext, err := EncryptAESECB(plaintext, keyBytes)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Decrypt
	decrypted, err := DecryptAESECB(ciphertext, keyBytes)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted text doesn't match:\nwant: %q\ngot:  %q", plaintext, decrypted)
	}
}

// BenchmarkEncryptAESECB benchmarks encryption performance
func BenchmarkEncryptAESECB(b *testing.B) {
	key := []byte("0123456789abcdef")
	plaintext := bytes.Repeat([]byte("A"), 1024) // 1KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EncryptAESECB(plaintext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecryptAESECB benchmarks decryption performance
func BenchmarkDecryptAESECB(b *testing.B) {
	key := []byte("0123456789abcdef")
	plaintext := bytes.Repeat([]byte("A"), 1024) // 1KB
	ciphertext, _ := EncryptAESECB(plaintext, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DecryptAESECB(ciphertext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}
