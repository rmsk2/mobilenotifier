package tools

import "fmt"

const VersionString = "1.2.0"

const MsgTextTomorrow = "Morgen"
const MsgTextToday = "Heute"
const MsgTextInSevenDays = "In 7 Tagen"

func GenerateNotificationText(prefix string, hour int, minute int, description string) string {
	return fmt.Sprintf("%s %02d:%02d %s", prefix, hour, minute, description)
}
