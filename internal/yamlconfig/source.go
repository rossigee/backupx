package yamlconfig

import "strconv"

type YAMLBackupSourceConf struct {
	Id              string `yaml:"id"`
	Name            string `yaml:"name"`
	Type            string `yaml:"type"`
	OtherAttributes map[string]string
}

func (c YAMLBackupSourceConf) GetId() string {
	return c.Id
}

func (c YAMLBackupSourceConf) GetName() string {
	return c.Name
}

func (c YAMLBackupSourceConf) GetType() string {
	return c.Type
}

func (c YAMLBackupSourceConf) GetOtherAttributes() map[string]string {
	return c.OtherAttributes
}

func parseSources(j []any) []YAMLBackupSourceConf {
	sources := []YAMLBackupSourceConf{}
	for _, a := range j {
		m, ok := a.(map[any]any)
		if !ok {
			continue
		}
		source := YAMLBackupSourceConf{}
		source.OtherAttributes = map[string]string{}
		for k, v := range m {
			key, ok := k.(string)
			if !ok {
				continue
			}
			switch key {
			case "id":
				if s, ok := v.(string); ok {
					source.Id = s
				}
			case "type":
				if s, ok := v.(string); ok {
					source.Type = s
				}
			case "name":
				if s, ok := v.(string); ok {
					source.Name = s
				}
			default:
				switch v := v.(type) {
				case string:
					source.OtherAttributes[key] = v
				case int:
					source.OtherAttributes[key] = strconv.Itoa(v)
				case float64:
					source.OtherAttributes[key] = strconv.FormatFloat(v, 'f', 6, 64)
				}
			}
		}
		sources = append(sources, source)
	}
	return sources
}
