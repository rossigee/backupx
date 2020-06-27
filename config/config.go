package config

import (
	"github.com/rossigee/backupx/destinations"
	"github.com/rossigee/backupx/notifications"
	"github.com/rossigee/backupx/sources"
)

type IBackupConfig interface {
	GetId() string
	GetName() string
	GetSourceConfigs() []sources.IBackupSourceConfig
	GetDestinationConfigs() []destinations.IBackupDestinationConfig
	GetNotificationConfigs() []notifications.IBackupNotificationConfig
}
