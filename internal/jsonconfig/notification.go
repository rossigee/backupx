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

func parseNotifications(j []interface{}) []JSONBackupNotificationConf {
	notifications := []JSONBackupNotificationConf{}
	for _, a := range j {
		notification := JSONBackupNotificationConf{}
		notification.OtherAttributes = map[string]string{}
		for k, v := range a.(map[string]interface{}) {
			switch k {
			case "id":
				notification.Id = v.(string)
			case "type":
				notification.Type = v.(string)
			case "name":
				notification.Name = v.(string)
			default:
				switch v.(type) {
				case string:
					notification.OtherAttributes[k] = v.(string)
				case float64:
					notification.OtherAttributes[k] = strconv.FormatFloat(v.(float64), 'f', 6, 64)
					// case int:
					// 	notification.OtherAttributes[k] = strconv.Itoa(v.(int))
				}
			}
		}
		notifications = append(notifications, notification)
	}
	return notifications
}
