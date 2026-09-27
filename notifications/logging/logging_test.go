package logging

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

func TestNewLoggingNotificationDefaults(t *testing.T) {
	config := testConfig{
		id:    "test",
		typ:   "logging",
		attrs: map[string]string{},
	}

	notif, err := NewLoggingNotification(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	if !notif.notifyOnSuccess {
		t.Error("Expected notifyOnSuccess to be true by default")
	}

	if !notif.notifyOnFailure {
		t.Error("Expected notifyOnFailure to be true by default")
	}
}

func TestLoggingNotificationSuccessOnly(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "logging",
		attrs: map[string]string{
			"notify_on_success": "1",
			"notify_on_failure": "0",
		},
	}

	notif, err := NewLoggingNotification(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	if !notif.notifyOnSuccess {
		t.Error("Expected notifyOnSuccess to be true")
	}

	if notif.notifyOnFailure {
		t.Error("Expected notifyOnFailure to be false")
	}
}

func TestLoggingNotificationSuccess(t *testing.T) {
	config := testConfig{
		id:    "test",
		typ:   "logging",
		attrs: map[string]string{},
	}

	notif, err := NewLoggingNotification(config)
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
}

func TestLoggingNotificationFailure(t *testing.T) {
	config := testConfig{
		id:    "test",
		typ:   "logging",
		attrs: map[string]string{},
	}

	notif, err := NewLoggingNotification(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	result := notifications.BackupResult{
		SourceId:   "test_source",
		SourceName: "Test Source",
		Success:    false,
		Error:      "Connection timeout",
		Message:    "Database unreachable",
		Duration:   1000,
	}

	err = notif.Notify(context.Background(), result)
	if err != nil {
		t.Fatalf("Failed to notify: %v", err)
	}
}
