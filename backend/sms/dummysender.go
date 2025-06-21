package sms

import (
	"fmt"
)

func NewDummySender() *dummySmsSender {
	res := new(dummySmsSender)

	martin := Recipient{
		DisplayName: displayMartin,
		Id:          idMartin,
		Address:     "SendSMS1",
	}

	push := Recipient{
		DisplayName: displayPush,
		Id:          idPush,
		Address:     "SendPush1",
	}

	res.recipientMap = map[string]Recipient{
		idMartin: martin,
		idPush:   push,
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
