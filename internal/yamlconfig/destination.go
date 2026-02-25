package yamlconfig

import "strconv"

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

func parseDestinations(j []any) []YAMLBackupDestinationConf {
	destinations := []YAMLBackupDestinationConf{}
	for _, a := range j {
		m, ok := a.(map[any]any)
		if !ok {
			continue
		}
		destination := YAMLBackupDestinationConf{}
		destination.OtherAttributes = map[string]string{}
		for k, v := range m {
			key, ok := k.(string)
			if !ok {
				continue
			}
			switch key {
			case "id":
				if s, ok := v.(string); ok {
					destination.Id = s
				}
			case "type":
				if s, ok := v.(string); ok {
					destination.Type = s
				}
			case "name":
				if s, ok := v.(string); ok {
					destination.Name = s
				}
			default:
				switch v := v.(type) {
				case string:
					destination.OtherAttributes[key] = v
				case int:
					destination.OtherAttributes[key] = strconv.Itoa(v)
				case float64:
					destination.OtherAttributes[key] = strconv.FormatFloat(v, 'f', 6, 64)
				}
			}
		}
		destinations = append(destinations, destination)
	}
	return destinations
}
