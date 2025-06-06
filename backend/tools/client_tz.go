package tools

import "time"

var clientTimeZone *time.Location = nil

func SetClientTZ(tz *time.Location) {
	clientTimeZone = tz
}

func SetDefaultTZ() {
	tz, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}

	clientTimeZone = tz
}

func ClientTZ() *time.Location {
	return clientTimeZone
}
