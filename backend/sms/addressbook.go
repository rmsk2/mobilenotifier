package sms

import (
	"encoding/json"
	"fmt"
	"notifier/repo"
)

type AddressSaver struct {
	recipientMap map[string]repo.Recipient
}

func NewAddressSaverFromJson(jsonData string) (*AddressSaver, error) {
	var m []repo.Recipient
	parsedRecipientMap := map[string]repo.Recipient{}

	err := json.Unmarshal([]byte(jsonData), &m)
	if err != nil {
		return nil, fmt.Errorf("unable to parse JSON recipient data: %v", err)
	}

	for _, j := range m {
		parsedRecipientMap[j.Id.String()] = j
	}

	return &AddressSaver{
		recipientMap: parsedRecipientMap,
	}, nil
}

func (a *AddressSaver) BBoltSave(db repo.AddrBookWrite) error {
	for _, v := range a.recipientMap {
		err := db.Upsert(&v)
		if err != nil {
			return fmt.Errorf("error saving address book: %v", err)
		}
	}

	return nil
}
