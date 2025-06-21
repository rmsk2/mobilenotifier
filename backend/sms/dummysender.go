package sms

import (
	"fmt"
)

func NewDummySender() *dummySmsSender {
	res := new(dummySmsSender)

	martin := Recipient{
		DisplayName: "martin",
		Id:          "0D69B617-12D0-4491-ADD8-D103CF3925A1",
		Address:     "SendSMS1",
	}

	push := Recipient{
		DisplayName: "push",
		Id:          "F55C84F3-A2C7-46DD-AF06-27AFF7FCCC16",
		Address:     "SendPush1",
	}

	res.recipientMap = map[string]Recipient{
		"0D69B617-12D0-4491-ADD8-D103CF3925A1": martin,
		"F55C84F3-A2C7-46DD-AF06-27AFF7FCCC16": push,
	}

	return res
}

type dummySmsSender struct {
	recipientMap map[string]Recipient
}

func (i *dummySmsSender) CheckRecipient(id string) (bool, error) {
	_, ok := i.recipientMap[id]
	return ok, nil
}

func (i *dummySmsSender) ListRecipients() ([]RecipientInfo, error) {
	info := []RecipientInfo{}

	for k := range i.recipientMap {
		i := RecipientInfo{
			Id:          k,
			DisplayName: i.recipientMap[k].DisplayName,
		}
		info = append(info, i)
	}

	return info, nil
}

func (i *dummySmsSender) Send(recipientId string, message string) error {
	v, ok := i.recipientMap[recipientId]
	if !ok {
		return fmt.Errorf("recipient %s is unknown", recipientId)
	}

	if len([]rune(message)) > lenMessageMax {
		temp := message
		message = string([]rune(temp)[:lenMessageMax])
	}

	fmt.Printf("Sending '%s' to '%s' using '%s'\n", message, v.DisplayName, v.Address)

	return nil
}
