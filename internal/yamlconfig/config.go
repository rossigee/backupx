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
	configs := make([]sources.IBackupSourceConfig, len(c.Sources))
	for i := range c.Sources {
		configs[i] = c.Sources[i]
	}
	return configs
}

func (c YAMLBackupConf) GetDestinationConfigs() []destinations.IBackupDestinationConfig {
	configs := make([]destinations.IBackupDestinationConfig, len(c.Destinations))
	for i := range c.Destinations {
		configs[i] = c.Destinations[i]
	}
	return configs
}

func (c YAMLBackupConf) GetNotificationConfigs() []notifications.IBackupNotificationConfig {
	configs := make([]notifications.IBackupNotificationConfig, len(c.Notifications))
	for i := range c.Notifications {
		configs[i] = c.Notifications[i]
	}
	return configs
}
