package jsonconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/rossigee/backupx/config"
)

func ParseJSONConfigFile(filename string) (config.IBackupConfig, error) {
	var conf JSONBackupConf

	rawJSON, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Error reading JSON file: %v", err)
	}

	var configMap map[string]interface{}
	err = json.Unmarshal(rawJSON, &configMap)
	if err != nil {
		return nil, fmt.Errorf("Error parsing JSON file: %v", err)
	}

	for k, v := range configMap {
		switch k {
		case "id":
			conf.Id = v.(string)
		case "name":
			conf.Name = v.(string)
		case "sources":
			conf.Sources = parseSources(v.([]interface{}))
		case "destinations":
			conf.Destinations = parseDestinations(v.([]interface{}))
		case "notifications":
			conf.Notifications = parseNotifications(v.([]interface{}))
		}
	}

	return conf, nil
}
