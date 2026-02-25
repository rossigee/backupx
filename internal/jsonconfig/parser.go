package jsonconfig

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rossigee/backupx/config"
)

func ParseJSONConfigFile(filename string) (config.IBackupConfig, error) {
	var conf JSONBackupConf

	rawJSON, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON file: %v", err)
	}

	var configMap map[string]any
	err = json.Unmarshal(rawJSON, &configMap)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON file: %v", err)
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
