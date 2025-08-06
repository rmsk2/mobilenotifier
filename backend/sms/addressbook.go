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
	CheckRecipient(r string) (bool, string, error)
	GetDefaultRecipientIds() []string
}

type AddressBook struct {
	recipientMap map[string]repo.Recipient
	senders      map[string]SmsSender
	defaultType  string
	defaultIds   []string
}

func NewAddressBookFromJson(jsonData string) (*AddressBook, error) {
	var m []repo.Recipient
	parsedRecipientMap := map[string]repo.Recipient{}
	defaults := []string{}

	err := json.Unmarshal([]byte(jsonData), &m)
	if err != nil {
		return nil, fmt.Errorf("unable to parse JSON recipient data: %v", err)
	}

	for _, j := range m {
		if j.IsDefault {
			defaults = append(defaults, j.Id.String())
		}
		parsedRecipientMap[j.Id.String()] = j
	}

	return &AddressBook{
		recipientMap: parsedRecipientMap,
		senders:      map[string]SmsSender{},
		defaultType:  TypeIFTTT,
		defaultIds:   defaults,
	}, nil
}

func (a *AddressBook) BBoltSave(db *repo.BBoltAddrBookRepo) error {
	for _, v := range a.recipientMap {
		err := db.Upsert(&v)
		if err != nil {
			return fmt.Errorf("error saving address book: %v", err)
		}
	}

	return nil
}

func (a *AddressBook) GetDefaultRecipientIds() []string {
	return a.defaultIds
}

func (a *AddressBook) SetDefaultRecipientIds(ids []string) {
	a.defaultIds = ids
}

func (a *AddressBook) SetDefaultType(t string) {
	a.defaultType = t
}

func (a *AddressBook) GetSender(recipientId string) SmsSender {
	var addrType string

	r, ok := a.recipientMap[recipientId]
	if !ok {
		addrType = a.defaultType
	} else {
		addrType = r.AddrType
	}

	sender, ok := a.senders[addrType]
	if !ok {
		sender = a.senders[a.defaultType]
	}

	return sender
}

func (a *AddressBook) ToJson() (string, error) {
	help := []repo.Recipient{}

	for _, j := range a.recipientMap {
		help = append(help, j)
	}

	data, err := json.Marshal(&help)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (a *AddressBook) AddRecipient(r repo.Recipient) {
	a.recipientMap[r.Id.String()] = r
}

func (a *AddressBook) AddSender(addrType string, s SmsSender) {
	a.senders[addrType] = s
}

func (a *AddressBook) CheckRecipient(r string) (bool, string, error) {
	recipient, ok := a.recipientMap[r]
	if !ok {
		return false, "", fmt.Errorf("recipient '%s' is unknown", r)
	}

	return ok, recipient.Address, nil

}

func (a *AddressBook) ListRecipients() ([]RecipientInfo, error) {
	info := []RecipientInfo{}

	for k := range a.recipientMap {
		i := RecipientInfo{
			Id:          k,
			DisplayName: a.recipientMap[k].DisplayName,
		}
		info = append(info, i)
	}

	return info, nil
}
