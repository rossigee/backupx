package jsonconfig

import "strconv"

type JSONBackupNotificationConf struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	OtherAttributes map[string]string
}

func (c JSONBackupNotificationConf) GetId() string {
	return c.Id
}

func (c JSONBackupNotificationConf) GetName() string {
	return c.Name
}

func (c JSONBackupNotificationConf) GetType() string {
	return c.Type
}

func (c JSONBackupNotificationConf) GetOtherAttributes() map[string]string {
	return c.OtherAttributes
}

func parseNotifications(j []any) []JSONBackupNotificationConf {
	notifications := []JSONBackupNotificationConf{}
	for _, a := range j {
		m, ok := a.(map[string]any)
		if !ok {
			continue
		}
		notification := JSONBackupNotificationConf{}
		notification.OtherAttributes = map[string]string{}
		for k, v := range m {
			switch k {
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
					notification.OtherAttributes[k] = v
				case float64:
					notification.OtherAttributes[k] = strconv.FormatFloat(v, 'f', 6, 64)
				}
			}
		}
		notifications = append(notifications, notification)
	}
	return notifications
}
