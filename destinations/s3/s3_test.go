package s3

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

type testConfig struct {
	id       string
	typ      string
	attrs    map[string]string
}

func (c testConfig) GetId() string {
	return c.id
}

func (c testConfig) GetType() string {
	return c.typ
}

func (c testConfig) GetName() string {
	return c.id
}

func (c testConfig) GetOtherAttributes() map[string]string {
	return c.attrs
}

func TestNewS3DestinationMissingBucket(t *testing.T) {
	config := testConfig{
		id:    "test",
		typ:   "s3",
		attrs: map[string]string{},
	}

	_, err := NewS3Destination(config, "testhost")
	if err == nil {
		t.Error("Expected error for missing bucket")
	}
}

func TestNewS3DestinationDefaults(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "s3",
		attrs: map[string]string{
			"bucket": "test-bucket",
		},
	}

	dest, err := NewS3Destination(config, "testhost")
	if err != nil {
		t.Fatalf("Failed to create S3 destination: %v", err)
	}

	if dest.bucket != "test-bucket" {
		t.Errorf("Expected bucket 'test-bucket', got '%s'", dest.bucket)
	}

	if dest.prefix != "backups" {
		t.Errorf("Expected default prefix 'backups', got '%s'", dest.prefix)
	}

	if dest.hostname != "testhost" {
		t.Errorf("Expected hostname 'testhost', got '%s'", dest.hostname)
	}
}

func TestNewS3DestinationWithCustomConfig(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "s3",
		attrs: map[string]string{
			"bucket":            "custom-bucket",
			"prefix":            "prod",
			"region":            "eu-west-1",
			"retention_copies":  "5",
			"retention_days":    "30",
			"endpoint_url":      "http://minio:9000",
		},
	}

	dest, err := NewS3Destination(config, "myhost")
	if err != nil {
		t.Fatalf("Failed to create S3 destination: %v", err)
	}

	if dest.bucket != "custom-bucket" {
		t.Errorf("Expected bucket 'custom-bucket', got '%s'", dest.bucket)
	}

	if dest.prefix != "prod" {
		t.Errorf("Expected prefix 'prod', got '%s'", dest.prefix)
	}

	if dest.retentionCopies != 5 {
		t.Errorf("Expected retention_copies 5, got %d", dest.retentionCopies)
	}

	if dest.retentionDays != 30 {
		t.Errorf("Expected retention_days 30, got %d", dest.retentionDays)
	}
}

func TestUploadPathGeneration(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "s3",
		attrs: map[string]string{
			"bucket": "test-bucket",
			"prefix": "backups",
		},
	}

	dest, err := NewS3Destination(config, "myhost")
	if err != nil {
		t.Fatalf("Failed to create S3 destination: %v", err)
	}

	testData := []byte("test backup data")
	reader := bytes.NewReader(testData)

	ctx := context.Background()

	path, err := dest.UploadReader(ctx, reader, "test.tar.gpg")
	if err != nil {
		if strings.Contains(err.Error(), "bucket does not exist") ||
			strings.Contains(err.Error(), "Access Denied") ||
			strings.Contains(err.Error(), "NoSuchBucket") {
			t.Skip("Skipping test that requires real S3 connection")
		}
		t.Fatalf("Failed to generate path: %v", err)
	}

	if !bytes.Contains([]byte(path), []byte("myhost")) {
		t.Errorf("Expected path to contain hostname 'myhost', got '%s'", path)
	}

	if !bytes.Contains([]byte(path), []byte("test.tar.gpg")) {
		t.Errorf("Expected path to contain filename 'test.tar.gpg', got '%s'", path)
	}
}
