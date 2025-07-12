package sms

import (
	"encoding/json"
	"fmt"
)

const TypeIFTTT = "IFTTT"
const TypeMail = "Mail"
const TypeDummy = "Dummy"

type Recipient struct {
	DisplayName string `json:"display_name"`
	Id          string `json:"id"`
	Address     string `json:"address"`
	AddrType    string `json:"addr_type"`
	IsDefault   bool   `json:"is_default"`
}

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
	recipientMap map[string]Recipient
	senders      map[string]SmsSender
	defaultType  string
	defaultIds   []string
}

func NewAddressBookFromJson(jsonData string) (*AddressBook, error) {
	var m []Recipient
	parsedRecipientMap := map[string]Recipient{}
	defaults := []string{}

	err := json.Unmarshal([]byte(jsonData), &m)
	if err != nil {
		return nil, fmt.Errorf("unable to parse JSON recipient data: %v", err)
	}

	for _, j := range m {
		if j.IsDefault {
			defaults = append(defaults, j.Id)
		}
		parsedRecipientMap[j.Id] = j
	}

	return &AddressBook{
		recipientMap: parsedRecipientMap,
		senders:      map[string]SmsSender{},
		defaultType:  TypeIFTTT,
		defaultIds:   defaults,
	}, nil
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
	help := []Recipient{}

	for _, j := range a.recipientMap {
		help = append(help, j)
	}

	data, err := json.Marshal(&help)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (a *AddressBook) AddRecipient(r Recipient) {
	a.recipientMap[r.Id] = r
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
