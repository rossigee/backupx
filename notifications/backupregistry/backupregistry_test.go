package backupregistry

import (
	"context"
	"testing"

	"github.com/rossigee/backupx/notifications"
)

type testConfig struct {
	id    string
	typ   string
	attrs map[string]string
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

func TestNewBackupRegistryMissingURL(t *testing.T) {
	config := testConfig{
		id:    "test",
		typ:   "backup-registry",
		attrs: map[string]string{},
	}

	_, err := NewBackupRegistry(config)
	if err == nil {
		t.Error("Expected error for missing URL")
	}
}

func TestNewBackupRegistryDefaults(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "backup-registry",
		attrs: map[string]string{
			"url": "https://backups.golder.tech/v1/backup-runs",
		},
	}

	br, err := NewBackupRegistry(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	if br.url != "https://backups.golder.tech/v1/backup-runs" {
		t.Errorf("Expected URL 'https://backups.golder.tech/v1/backup-runs', got '%s'", br.url)
	}

	if br.token != "" {
		t.Errorf("Expected empty token, got '%s'", br.token)
	}

	if len(br.metadata) != 0 {
		t.Errorf("Expected empty metadata, got %d items", len(br.metadata))
	}
}

func TestNewBackupRegistryWithToken(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "backup-registry",
		attrs: map[string]string{
			"url":   "https://backups.golder.tech/v1/backup-runs",
			"token": "secret-token-123",
		},
	}

	br, err := NewBackupRegistry(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	if br.token != "secret-token-123" {
		t.Errorf("Expected token 'secret-token-123', got '%s'", br.token)
	}
}

func TestNewBackupRegistryWithMetadata(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "backup-registry",
		attrs: map[string]string{
			"url":         "https://backups.golder.tech/v1/backup-runs",
			"environment": "production",
			"region":      "us-east-1",
		},
	}

	br, err := NewBackupRegistry(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	if br.metadata["environment"] != "production" {
		t.Errorf("Expected environment 'production', got '%s'", br.metadata["environment"])
	}

	if br.metadata["region"] != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got '%s'", br.metadata["region"])
	}
}

func TestNotifySuccess(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "backup-registry",
		attrs: map[string]string{
			"url": "https://invalid-url-for-testing.example.com",
		},
	}

	br, err := NewBackupRegistry(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	result := notifications.BackupResult{
		SourceId:   "my-db",
		SourceName: "Production Database",
		Success:    true,
		Duration:   5000,
		BackupPath: "s3://bucket/backup-2026-09-27.sql.gpg",
	}

	err = br.Notify(context.Background(), result)
	if err == nil {
		t.Error("Expected connection error, but notification was created successfully")
	}
}

func TestNotifyFailure(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "backup-registry",
		attrs: map[string]string{
			"url": "https://invalid-url-for-testing.example.com",
		},
	}

	br, err := NewBackupRegistry(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	result := notifications.BackupResult{
		SourceId:   "my-db",
		SourceName: "Production Database",
		Success:    false,
		Duration:   1000,
		Error:      "Connection timeout",
	}

	err = br.Notify(context.Background(), result)
	if err == nil {
		t.Error("Expected connection error, but notification was created successfully")
	}
}
