package jsonconfig

import "strconv"

type JSONBackupDestinationConf struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	OtherAttributes map[string]string
}

func (c JSONBackupDestinationConf) GetId() string {
	return c.Id
}

func (c JSONBackupDestinationConf) GetName() string {
	return c.Name
}

func (c JSONBackupDestinationConf) GetType() string {
	return c.Type
}

func (c JSONBackupDestinationConf) GetOtherAttributes() map[string]string {
	return c.OtherAttributes
}

func parseDestinations(j []any) []JSONBackupDestinationConf {
	destinations := []JSONBackupDestinationConf{}
	for _, a := range j {
		m, ok := a.(map[string]any)
		if !ok {
			continue
		}
		destination := JSONBackupDestinationConf{}
		destination.OtherAttributes = map[string]string{}
		for k, v := range m {
			switch k {
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
					destination.OtherAttributes[k] = v
				case float64:
					destination.OtherAttributes[k] = strconv.FormatFloat(v, 'f', 6, 64)
				}
			}
		}
		destinations = append(destinations, destination)
	}
	return destinations
}
