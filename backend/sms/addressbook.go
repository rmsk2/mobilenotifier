package sms

import (
	"encoding/json"
	"fmt"
	"notifier/repo"
)

const TypeIFTTT = "IFTTT"
const TypeMail = "Mail"
const TypeDummy = "Dummy"

type RecipientInfo struct {
	Id          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type SmsAddressBook interface {
	ListRecipients() ([]RecipientInfo, error)
	GetSender(addrType string) SmsSender
	AddSender(addrType string, s SmsSender)
	SetDefaultType(t string)
	CheckRecipient(r string) (bool, string, error)
	GetDefaultRecipientIds() []string
}

type AddressBook struct {
	recipientMap map[string]repo.Recipient
}

func NewAddressBookFromJson(jsonData string) (*AddressBook, error) {
	var m []repo.Recipient
	parsedRecipientMap := map[string]repo.Recipient{}

	err := json.Unmarshal([]byte(jsonData), &m)
	if err != nil {
		return nil, fmt.Errorf("unable to parse JSON recipient data: %v", err)
	}

	for _, j := range m {
		parsedRecipientMap[j.Id.String()] = j
	}

	return &AddressBook{
		recipientMap: parsedRecipientMap,
	}, nil
}

func (a *AddressBook) BBoltSave(db repo.AddrBookWrite) error {
	for _, v := range a.recipientMap {
		err := db.Upsert(&v)
		if err != nil {
			return fmt.Errorf("error saving address book: %v", err)
		}
	}

	return nil
}
