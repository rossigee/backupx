package notifications

import "context"

type IBackupNotificationConfig interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}

type BackupResult struct {
	SourceId   string
	SourceName string
	Success    bool
	Error      string
	Message    string
	Duration   int64 // milliseconds
	BackupPath string
}

type IBackupNotification interface {
	Notify(ctx context.Context, result BackupResult) error
}
