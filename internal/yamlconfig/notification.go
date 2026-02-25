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

func parseNotifications(j []any) []YAMLBackupNotificationConf {
	notifications := []YAMLBackupNotificationConf{}
	for _, a := range j {
		m, ok := a.(map[any]any)
		if !ok {
			continue
		}
		notification := YAMLBackupNotificationConf{}
		notification.OtherAttributes = map[string]string{}
		for k, v := range m {
			key, ok := k.(string)
			if !ok {
				continue
			}
			switch key {
			case "id":
				if s, ok := v.(string); ok {
					notification.Id = s
				}
			case "type":
				if s, ok := v.(string); ok {
					notification.Type = s
				}
			case "name":
				if s, ok := v.(string); ok {
					notification.Name = s
				}
			default:
				switch v := v.(type) {
				case string:
					notification.OtherAttributes[key] = v
				case int:
					notification.OtherAttributes[key] = strconv.Itoa(v)
				case float64:
					notification.OtherAttributes[key] = strconv.FormatFloat(v, 'f', 6, 64)
				}
			}
		}
		notifications = append(notifications, notification)
	}
	return notifications
}
