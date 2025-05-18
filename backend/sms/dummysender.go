package sms

import (
	"fmt"
)

func NewDummySender() SmsSender {
	res := new(dummySmsSender)
	res.recipientMap = map[string]string{
		"martin": "SendSMS1",
	}

	return res
}

type dummySmsSender struct {
	recipientMap map[string]string
}

func (i *dummySmsSender) CheckRecipient(r string) (bool, error) {
	_, ok := i.recipientMap[r]
	return ok, nil
}

func (i *dummySmsSender) ListRecipients() ([]string, error) {
	keys := make([]string, len(i.recipientMap))

	for k := range i.recipientMap {
		keys = append(keys, k)
	}

	return keys, nil
}

func (i *dummySmsSender) Send(recipient string, message string) error {
	v, ok := i.recipientMap[recipient]
	if !ok {
		return fmt.Errorf("recipient %s is unknown", recipient)
	}

	fmt.Printf("Sending '%s' to '%s' using '%s'\n", message, recipient, v)

	return nil
}
