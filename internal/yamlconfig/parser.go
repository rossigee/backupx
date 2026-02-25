package yamlconfig

import (
	"fmt"
	"os"

	"github.com/rossigee/backupx/config"
	"gopkg.in/yaml.v2"
)

func ParseYAMLConfigFile(filename string) (config.IBackupConfig, error) {
	var conf YAMLBackupConf

	rawYAML, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	var configMap map[string]any
	err = yaml.Unmarshal(rawYAML, &configMap)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML file: %v", err)
	}

	for k, v := range configMap {
		switch k {
		case "id":
			if s, ok := v.(string); ok {
				conf.Id = s
			}
		case "name":
			if s, ok := v.(string); ok {
				conf.Name = s
			}
		case "sources":
			if slice, ok := v.([]any); ok {
				conf.Sources = parseSources(slice)
			}
		case "destinations":
			if slice, ok := v.([]any); ok {
				conf.Destinations = parseDestinations(slice)
			}
		case "notifications":
			if slice, ok := v.([]any); ok {
				conf.Notifications = parseNotifications(slice)
			}
		}
	}

	return conf, nil
}
