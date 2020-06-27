package yamlconfig

import "strconv"

type YAMLBackupNotificationConf struct {
	Id              string `yaml:"id"`
	Name            string `yaml:"name"`
	Type            string `yaml:"type"`
	OtherAttributes map[string]string
}

func (c YAMLBackupNotificationConf) GetId() string {
	return c.Id
}

func (c YAMLBackupNotificationConf) GetName() string {
	return c.Name
}

func (c YAMLBackupNotificationConf) GetType() string {
	return c.Type
}

func (c YAMLBackupNotificationConf) GetOtherAttributes() map[string]string {
	return c.OtherAttributes
}

func parseNotifications(j []interface{}) []YAMLBackupNotificationConf {
	notifications := []YAMLBackupNotificationConf{}
	for _, a := range j {
		notification := YAMLBackupNotificationConf{}
		notification.OtherAttributes = map[string]string{}
		for k, v := range a.(map[interface{}]interface{}) {
			switch k.(string) {
			case "id":
				notification.Id = v.(string)
			case "type":
				notification.Type = v.(string)
			case "name":
				notification.Name = v.(string)
			default:
				switch v.(type) {
				case string:
					notification.OtherAttributes[k.(string)] = v.(string)
				case float64:
					notification.OtherAttributes[k.(string)] = strconv.FormatFloat(v.(float64), 'f', 6, 64)
					// case int:
					// 	notification.OtherAttributes[k] = strconv.Itoa(v.(int))
				}
			}
		}
		notifications = append(notifications, notification)
	}
	return notifications
}
