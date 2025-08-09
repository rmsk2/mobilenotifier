package sms

import (
	"encoding/json"
	"fmt"
	"notifier/repo"
)

type AddressSaver struct {
	recipients []repo.Recipient
}

func NewAddressSaverFromJson(jsonData string) (*AddressSaver, error) {
	var m []repo.Recipient

	err := json.Unmarshal([]byte(jsonData), &m)
	if err != nil {
		return nil, fmt.Errorf("unable to parse JSON recipient data: %v", err)
	}

	return &AddressSaver{
		recipients: m,
	}, nil
}

func (a *AddressSaver) BBoltSave(db repo.AddrBookWrite) error {
	for _, v := range a.recipients {
		err := db.Upsert(&v)
		if err != nil {
			return fmt.Errorf("error saving address book: %v", err)
		}
	}

	return nil
}
