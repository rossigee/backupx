package yamlconfig

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

func parseSources(j []interface{}) []YAMLBackupSourceConf {
	sources := []YAMLBackupSourceConf{}
	for _, a := range j {
		source := YAMLBackupSourceConf{}
		source.OtherAttributes = map[string]string{}
		for k, v := range a.(map[interface{}]interface{}) {
			switch k.(string) {
			case "id":
				source.Id = v.(string)
			case "type":
				source.Type = v.(string)
			case "name":
				source.Name = v.(string)
			default:
				source.OtherAttributes[k.(string)] = v.(string)
			}
		}
		sources = append(sources, source)
	}
	return sources
}
