package sms

import "fmt"

const displayMartin = "Martin via SMS"
const displayPush = "Pushmessage"
const idMartin = "0D69B617-12D0-4491-ADD8-D103CF3925A1"
const idPush = "F55C84F3-A2C7-46DD-AF06-27AFF7FCCC16"
const addrSMS = "SendSMS1"
const addrPush = "SendPush1"

const TypeSMS = "SMS"
const TypeMail = "Mail"
const TypeDummy = "Dummy"

type Recipient struct {
	DisplayName string
	Id          string
	Address     string
	AddrType    string
}

type RecipientInfo struct {
	Id          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type SmsAddressBook interface {
	ListRecipients() ([]RecipientInfo, error)
	GetSender(addrType string) SmsSender
	CheckRecipient(r string) (bool, string, error)
}

type AddressBook struct {
	recipientMap map[string]Recipient
	senders      map[string]SmsSender
	defaultType  string
}

func makeRecipientMap() map[string]Recipient {
	martin := Recipient{
		DisplayName: displayMartin,
		Id:          idMartin,
		Address:     addrSMS,
		AddrType:    TypeSMS,
	}

	push := Recipient{
		DisplayName: displayPush,
		Id:          idPush,
		Address:     addrPush,
		AddrType:    TypeSMS,
	}

	res := map[string]Recipient{
		idMartin: martin,
		idPush:   push,
	}

	return res
}

func NewAddressBook() *AddressBook {
	return &AddressBook{
		recipientMap: makeRecipientMap(),
		senders:      map[string]SmsSender{},
		defaultType:  TypeSMS,
	}
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
