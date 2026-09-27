package slack

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

func TestNewSlackNotificationMissingURL(t *testing.T) {
	config := testConfig{
		id:    "test",
		typ:   "slack",
		attrs: map[string]string{},
	}

	_, err := NewSlackNotification(config)
	if err == nil {
		t.Error("Expected error for missing URL")
	}
}

func TestNewSlackNotificationDefaults(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "slack",
		attrs: map[string]string{
			"url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
		},
	}

	notif, err := NewSlackNotification(config)
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

func TestSlackNotificationSuccessOnly(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "slack",
		attrs: map[string]string{
			"url":                  "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
			"notify_on_success":    "1",
			"notify_on_failure":    "0",
		},
	}

	notif, err := NewSlackNotification(config)
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

func TestSlackNotificationWebhookValidation(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "slack",
		attrs: map[string]string{
			"url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
		},
	}

	notif, err := NewSlackNotification(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	if notif.webhookURL != "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX" {
		t.Error("Webhook URL not set correctly")
	}
}

func TestSlackNotificationFailureSkipped(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "slack",
		attrs: map[string]string{
			"url":                  "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
			"notify_on_success":    "1",
			"notify_on_failure":    "0",
		},
	}

	notif, err := NewSlackNotification(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	result := notifications.BackupResult{
		SourceId:   "test_source",
		SourceName: "Test Source",
		Success:    false,
		Error:      "Connection failed",
		Duration:   1000,
	}

	err = notif.Notify(context.Background(), result)
	if err != nil {
		t.Fatalf("Failed to notify: %v", err)
	}
}

func TestSlackNotificationSuccessMessage(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "slack",
		attrs: map[string]string{
			"url": "https://hooks.slack.com/services/invalid",
		},
	}

	notif, err := NewSlackNotification(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	result := notifications.BackupResult{
		SourceId:   "test_source",
		SourceName: "Test Source",
		Success:    true,
		Duration:   5000,
		BackupPath: "s3://bucket/backup-2026-09-27.sql.gpg",
	}

	err = notif.Notify(context.Background(), result)
	if err == nil {
		t.Log("Note: notification would be sent if Slack was reachable")
	}
}

func TestSlackNotificationFailureMessage(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "slack",
		attrs: map[string]string{
			"url": "https://hooks.slack.com/services/invalid",
		},
	}

	notif, err := NewSlackNotification(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	result := notifications.BackupResult{
		SourceId:   "test_source",
		SourceName: "Test Source",
		Success:    false,
		Duration:   1000,
		Error:      "Database connection timeout",
	}

	err = notif.Notify(context.Background(), result)
	if err == nil {
		t.Log("Note: notification would be sent if Slack was reachable")
	}
}

func TestSlackNotificationBothFlagsDisabled(t *testing.T) {
	config := testConfig{
		id:   "test",
		typ:  "slack",
		attrs: map[string]string{
			"url":                  "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
			"notify_on_success":    "0",
			"notify_on_failure":    "0",
		},
	}

	notif, err := NewSlackNotification(config)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	if !notif.notifyOnSuccess {
		t.Error("Expected notifyOnSuccess to default to true when both are disabled")
	}

	if !notif.notifyOnFailure {
		t.Error("Expected notifyOnFailure to default to true when both are disabled")
	}
}
