package logging

import (
	"context"
	"fmt"
	"log"

	"github.com/rossigee/backupx/notifications"
)

type LoggingNotification struct {
	notifyOnSuccess bool
	notifyOnFailure bool
}

type NotificationConfig interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}

func NewLoggingNotification(config NotificationConfig) (*LoggingNotification, error) {
	attrs := config.GetOtherAttributes()

	notifyOnSuccess := attrs["notify_on_success"] != "0" && attrs["notify_on_success"] != ""
	notifyOnFailure := attrs["notify_on_failure"] != "0" && attrs["notify_on_failure"] != ""

	if !notifyOnSuccess && !notifyOnFailure {
		notifyOnSuccess = true
		notifyOnFailure = true
	}

	return &LoggingNotification{
		notifyOnSuccess: notifyOnSuccess,
		notifyOnFailure: notifyOnFailure,
	}, nil
}

func (l *LoggingNotification) Notify(ctx context.Context, result notifications.BackupResult) error {
	shouldNotify := (result.Success && l.notifyOnSuccess) || (!result.Success && l.notifyOnFailure)

	if !shouldNotify {
		return nil
	}

	status := "SUCCESS"
	if !result.Success {
		status = "FAILURE"
	}

	message := fmt.Sprintf("Backup %s: %s (%s) - Duration: %dms",
		status,
		result.SourceId,
		result.SourceName,
		result.Duration,
	)

	if result.Success {
		message += fmt.Sprintf(" - Path: %s", result.BackupPath)
	} else {
		message += fmt.Sprintf(" - Error: %s", result.Error)
		if result.Message != "" {
			message += fmt.Sprintf(" (%s)", result.Message)
		}
	}

	log.Println(message)
	return nil
}
