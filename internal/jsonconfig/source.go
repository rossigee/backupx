package jsonconfig

type JSONBackupSourceConf struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	OtherAttributes map[string]string
}

func (c JSONBackupSourceConf) GetId() string {
	return c.Id
}

func (c JSONBackupSourceConf) GetName() string {
	return c.Name
}

func (c JSONBackupSourceConf) GetType() string {
	return c.Type
}

func (c JSONBackupSourceConf) GetOtherAttributes() map[string]string {
	return c.OtherAttributes
}

func parseSources(j []interface{}) []JSONBackupSourceConf {
	sources := []JSONBackupSourceConf{}
	for _, a := range j {
		source := JSONBackupSourceConf{}
		source.OtherAttributes = map[string]string{}
		for k, v := range a.(map[string]interface{}) {
			switch k {
			case "id":
				source.Id = v.(string)
			case "type":
				source.Type = v.(string)
			case "name":
				source.Name = v.(string)
			default:
				source.OtherAttributes[k] = v.(string)
			}
		}
		sources = append(sources, source)
	}
	return sources
}
