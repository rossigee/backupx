package jsonconfig

import (
	"github.com/rossigee/backupx/destinations"
	"github.com/rossigee/backupx/notifications"
	"github.com/rossigee/backupx/sources"
)

type JSONBackupConf struct {
	Id            string                       `json:"id"`
	Name          string                       `json:"name"`
	Sources       []JSONBackupSourceConf       `json:"sources"`
	Destinations  []JSONBackupDestinationConf  `json:"destinations"`
	Notifications []JSONBackupNotificationConf `json:"notifications"`
}

func (c JSONBackupConf) GetId() string {
	return c.Id
}

func (c JSONBackupConf) GetName() string {
	return c.Name
}

func (c JSONBackupConf) GetSourceConfigs() []sources.IBackupSourceConfig {
	var configs []sources.IBackupSourceConfig
	configs = make([]sources.IBackupSourceConfig, len(c.Sources))
	for i := 0; i < len(c.Sources); i++ {
		configs[i] = c.Sources[i]
	}
	return configs
}

func (c JSONBackupConf) GetDestinationConfigs() []destinations.IBackupDestinationConfig {
	var configs []destinations.IBackupDestinationConfig
	configs = make([]destinations.IBackupDestinationConfig, len(c.Destinations))
	for i := 0; i < len(c.Destinations); i++ {
		configs[i] = c.Destinations[i]
	}
	return configs
}

func (c JSONBackupConf) GetNotificationConfigs() []notifications.IBackupNotificationConfig {
	var configs []notifications.IBackupNotificationConfig
	configs = make([]notifications.IBackupNotificationConfig, len(c.Notifications))
	for i := 0; i < len(c.Notifications); i++ {
		configs[i] = c.Notifications[i]
	}
	return configs
}
