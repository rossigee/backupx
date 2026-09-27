package flagfile

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rossigee/backupx/notifications"
)

type FlagFileNotification struct {
	filepath string
}

type NotificationConfig interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}

func NewFlagFileNotification(config NotificationConfig) (*FlagFileNotification, error) {
	attrs := config.GetOtherAttributes()
	filepath := attrs["filepath"]
	if filepath == "" {
		filepath = attrs["flagfile"]
	}
	if filepath == "" {
		return nil, fmt.Errorf("flagfile notification missing 'filepath' or 'flagfile' attribute")
	}

	return &FlagFileNotification{
		filepath: filepath,
	}, nil
}

func (f *FlagFileNotification) Notify(ctx context.Context, result notifications.BackupResult) error {
	if !result.Success {
		return nil
	}

	content := fmt.Sprintf("Backup completed at %s\nSource: %s (%s)\nPath: %s\nDuration: %dms\n",
		time.Now().Format(time.RFC3339),
		result.SourceId,
		result.SourceName,
		result.BackupPath,
		result.Duration,
	)

	err := os.WriteFile(f.filepath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write flagfile: %v", err)
	}

	return nil
}
