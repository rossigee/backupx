package yamlconfig

import (
	"github.com/rossigee/backupx/destinations"
	"github.com/rossigee/backupx/notifications"
	"github.com/rossigee/backupx/sources"
)

type YAMLBackupConf struct {
	Id            string                       `yaml:"id"`
	Name          string                       `yaml:"name"`
	Sources       []YAMLBackupSourceConf       `yaml:"sources"`
	Destinations  []YAMLBackupDestinationConf  `yaml:"destinations"`
	Notifications []YAMLBackupNotificationConf `yaml:"notifications"`
}

func (c YAMLBackupConf) GetId() string {
	return c.Id
}

func (c YAMLBackupConf) GetName() string {
	return c.Name
}

func (c YAMLBackupConf) GetSourceConfigs() []sources.IBackupSourceConfig {
	var configs []sources.IBackupSourceConfig
	configs = make([]sources.IBackupSourceConfig, len(c.Sources))
	for i := 0; i < len(c.Sources); i++ {
		configs[i] = c.Sources[i]
	}
	return configs
}

func (c YAMLBackupConf) GetDestinationConfigs() []destinations.IBackupDestinationConfig {
	var configs []destinations.IBackupDestinationConfig
	configs = make([]destinations.IBackupDestinationConfig, len(c.Destinations))
	for i := 0; i < len(c.Destinations); i++ {
		configs[i] = c.Destinations[i]
	}
	return configs
}

func (c YAMLBackupConf) GetNotificationConfigs() []notifications.IBackupNotificationConfig {
	var configs []notifications.IBackupNotificationConfig
	configs = make([]notifications.IBackupNotificationConfig, len(c.Notifications))
	for i := 0; i < len(c.Notifications); i++ {
		configs[i] = c.Notifications[i]
	}
	return configs
}
