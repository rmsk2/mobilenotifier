package tools

import (
	"encoding/json"
	"fmt"
)

const VersionString = "1.5.5"

const MsgTextTomorrow = "Morgen"
const MsgTextToday = "Heute"
const MsgTextInSevenDays = "In 7 Tagen"

type JsonNotification struct {
	Prefix  string `json:"prefix"`
	Hour    int    `json:"hour"`
	Minute  int    `json:"minute"`
	Message string `json:"message"`
}

func GenerateNotificationText(prefix string, hour int, minute int, description string) string {
	return fmt.Sprintf("%s %02d:%02d %s", prefix, hour, minute, description)
}

func GenerateNotificationTextNoTimestamp(prefix string, hour int, minute int, description string) string {
	return description
}

func GenerateNotificationTextJson(prefix string, hour int, minute int, description string) string {
	jNot := JsonNotification{
		Prefix:  prefix,
		Hour:    hour,
		Minute:  minute,
		Message: description,
	}

	data, _ := json.Marshal(&jNot)

	return string(data)
}
