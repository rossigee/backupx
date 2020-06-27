package yamlconfig

import (
	"strconv"
)

type YAMLBackupDestinationConf struct {
	Id              string `yaml:"id"`
	Name            string `yaml:"name"`
	Type            string `yaml:"type"`
	OtherAttributes map[string]string
}

func (c YAMLBackupDestinationConf) GetId() string {
	return c.Id
}

func (c YAMLBackupDestinationConf) GetName() string {
	return c.Name
}

func (c YAMLBackupDestinationConf) GetType() string {
	return c.Type
}

func (c YAMLBackupDestinationConf) GetOtherAttributes() map[string]string {
	return c.OtherAttributes
}

func parseDestinations(j []interface{}) []YAMLBackupDestinationConf {
	destinations := []YAMLBackupDestinationConf{}
	for _, a := range j {
		destination := YAMLBackupDestinationConf{}
		destination.OtherAttributes = map[string]string{}
		for k, v := range a.(map[interface{}]interface{}) {
			switch k.(string) {
			case "id":
				destination.Id = v.(string)
			case "type":
				destination.Type = v.(string)
			case "name":
				destination.Name = v.(string)
			default:
				switch v.(type) {
				case string:
					destination.OtherAttributes[k.(string)] = v.(string)
				case float64:
					destination.OtherAttributes[k.(string)] = strconv.FormatFloat(v.(float64), 'f', 6, 64)
					// case int:
					// 	destination.OtherAttributes[k] = strconv.Itoa(v.(int))
				}
			}
		}
		destinations = append(destinations, destination)
	}
	return destinations
}
