package sms

import (
	"fmt"
)

func NewDummySender() *dummySmsSender {
	res := new(dummySmsSender)
	res.recipientMap = makeRecipientMap()
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
	return listRecipientsOnMap(i.recipientMap), nil
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
