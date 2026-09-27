package flagfile

import (
	"context"
	"os"
	"path/filepath"
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

func TestNewFlagFileNotificationMissingPath(t *testing.T) {
	config := testConfig{
		id:    "test",
		typ:   "flagfile",
		attrs: map[string]string{},
	}

	_, err := NewFlagFileNotification(config)
	if err == nil {
		t.Error("Expected error for missing filepath")
	}
}

func TestFlagFileNotificationSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	flagPath := filepath.Join(tmpDir, "backup.ok")

	config := testConfig{
		id:   "test",
		typ:  "flagfile",
		attrs: map[string]string{
			"filepath": flagPath,
		},
	}

	notif, err := NewFlagFileNotification(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	result := notifications.BackupResult{
		SourceId:   "test_source",
		SourceName: "Test Source",
		Success:    true,
		Message:    "Backup succeeded",
		Duration:   5000,
		BackupPath: "s3://bucket/backup.tar.gpg",
	}

	err = notif.Notify(context.Background(), result)
	if err != nil {
		t.Fatalf("Failed to notify: %v", err)
	}

	if _, err := os.Stat(flagPath); err != nil {
		t.Fatalf("Flagfile not created: %v", err)
	}

	content, err := os.ReadFile(flagPath)
	if err != nil {
		t.Fatalf("Failed to read flagfile: %v", err)
	}

	if len(content) == 0 {
		t.Error("Flagfile is empty")
	}
}

func TestFlagFileNotificationFailureIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	flagPath := filepath.Join(tmpDir, "backup.ok")

	config := testConfig{
		id:   "test",
		typ:  "flagfile",
		attrs: map[string]string{
			"filepath": flagPath,
		},
	}

	notif, err := NewFlagFileNotification(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	result := notifications.BackupResult{
		SourceId:   "test_source",
		SourceName: "Test Source",
		Success:    false,
		Error:      "Backup failed",
		Message:    "Connection timeout",
		Duration:   1000,
	}

	err = notif.Notify(context.Background(), result)
	if err != nil {
		t.Fatalf("Failed to notify: %v", err)
	}

	if _, err := os.Stat(flagPath); err == nil {
		t.Error("Flagfile should not be created on failure")
	}
}
