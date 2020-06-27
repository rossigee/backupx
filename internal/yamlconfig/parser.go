package yamlconfig

import (
	"fmt"
	"io/ioutil"

	"github.com/rossigee/backupx/config"
	"gopkg.in/yaml.v2"
)

func ParseYAMLConfigFile(filename string) (config.IBackupConfig, error) {
	var conf YAMLBackupConf

	rawYAML, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Error reading YAML file: %v", err)
	}

	var configMap map[string]interface{}
	err = yaml.Unmarshal(rawYAML, &configMap)
	if err != nil {
		return nil, fmt.Errorf("Error parsing YAML file: %v", err)
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
