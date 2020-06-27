package jsonconfig

import (
	"strconv"
)

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

func parseDestinations(j []interface{}) []JSONBackupDestinationConf {
	destinations := []JSONBackupDestinationConf{}
	for _, a := range j {
		destination := JSONBackupDestinationConf{}
		destination.OtherAttributes = map[string]string{}
		for k, v := range a.(map[string]interface{}) {
			switch k {
			case "id":
				destination.Id = v.(string)
			case "type":
				destination.Type = v.(string)
			case "name":
				destination.Name = v.(string)
			default:
				switch v.(type) {
				case string:
					destination.OtherAttributes[k] = v.(string)
				case float64:
					destination.OtherAttributes[k] = strconv.FormatFloat(v.(float64), 'f', 6, 64)
					// case int:
					// 	destination.OtherAttributes[k] = strconv.Itoa(v.(int))
				}
			}
		}
		destinations = append(destinations, destination)
	}
	return destinations
}
