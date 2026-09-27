package encryption

import (
	"bytes"
	"io"
	"testing"
)

func TestNewGPGEncryptor(t *testing.T) {
	tests := []struct {
		name         string
		passphrase   string
		compressOnly bool
	}{
		{
			name:         "with_passphrase_and_encryption",
			passphrase:   "secret123",
			compressOnly: false,
		},
		{
			name:         "with_passphrase_compression_only",
			passphrase:   "secret123",
			compressOnly: true,
		},
		{
			name:         "no_passphrase_compression_only",
			passphrase:   "",
			compressOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := NewGPGEncryptor(tt.passphrase, tt.compressOnly)
			if enc.passphrase != tt.passphrase {
				t.Errorf("Expected passphrase '%s', got '%s'", tt.passphrase, enc.passphrase)
			}
			if enc.compressOnly != tt.compressOnly {
				t.Errorf("Expected compressOnly %v, got %v", tt.compressOnly, enc.compressOnly)
			}
		})
	}
}

func TestEncryptMissingPassphraseWithEncryption(t *testing.T) {
	enc := NewGPGEncryptor("", false)
	data := []byte("test data")
	reader := bytes.NewReader(data)

	_, err := enc.Encrypt(reader)
	if err == nil {
		t.Error("Expected error for missing passphrase with encryption")
	}
}

func TestEncryptCompressionOnly(t *testing.T) {
	enc := NewGPGEncryptor("", true)
	largeData := bytes.Repeat([]byte("test data to compress "), 100)
	reader := bytes.NewReader(largeData)

	compressed, err := enc.Encrypt(reader)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	defer func() {
		_ = compressed.Close()
	}()

	result, err := io.ReadAll(compressed)
	if err != nil {
		t.Fatalf("Failed to read compressed data: %v", err)
	}

	if len(result) < len(largeData)/2 && len(result) > 0 {
		t.Log("Compression working: input", len(largeData), "output", len(result))
	}
}

func TestEncryptWithPassphrase(t *testing.T) {
	enc := NewGPGEncryptor("testpass", false)
	data := []byte("test data to encrypt")
	reader := bytes.NewReader(data)

	encrypted, err := enc.Encrypt(reader)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	defer func() {
		_ = encrypted.Close()
	}()

	result, err := io.ReadAll(encrypted)
	if err != nil {
		t.Fatalf("Failed to read encrypted data: %v", err)
	}

	if len(result) == 0 {
		t.Error("Encrypted data is empty")
	}

	if bytes.Equal(result, data) {
		t.Error("Encrypted data should differ from input")
	}

	if !bytes.Contains(result, []byte("PK")) && !bytes.HasPrefix(result, []byte("\x8c")) {
		t.Log("Warning: encrypted data doesn't match expected GPG format")
	}
}

func TestEncryptLargeData(t *testing.T) {
	enc := NewGPGEncryptor("testpass", false)

	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	reader := bytes.NewReader(largeData)
	encrypted, err := enc.Encrypt(reader)
	if err != nil {
		t.Fatalf("Encrypt failed for large data: %v", err)
	}
	defer func() {
		_ = encrypted.Close()
	}()

	result, err := io.ReadAll(encrypted)
	if err != nil {
		t.Fatalf("Failed to read encrypted data: %v", err)
	}

	if len(result) == 0 {
		t.Error("Encrypted data is empty")
	}
}

func TestGPGReaderClose(t *testing.T) {
	enc := NewGPGEncryptor("testpass", false)
	data := []byte("test data")
	reader := bytes.NewReader(data)

	encrypted, err := enc.Encrypt(reader)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	err = encrypted.Close()
	if err != nil {
		t.Logf("Close error (expected): %v", err)
	}
}

func TestEncryptEmptyInput(t *testing.T) {
	enc := NewGPGEncryptor("testpass", false)
	reader := bytes.NewReader([]byte{})

	encrypted, err := enc.Encrypt(reader)
	if err != nil {
		t.Fatalf("Encrypt failed for empty input: %v", err)
	}
	defer func() {
		_ = encrypted.Close()
	}()

	result, err := io.ReadAll(encrypted)
	if err != nil {
		t.Fatalf("Failed to read encrypted data: %v", err)
	}

	if len(result) == 0 {
		t.Log("Note: empty input produces empty encryption (expected)")
	}
}
