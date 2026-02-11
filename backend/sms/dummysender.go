package sms

import (
	"fmt"
)

func NewDummySender() *dummySmsSender {
	res := new(dummySmsSender)
	return res
}

type dummySmsSender struct {
}

func (i *dummySmsSender) GetName() string {
	return "IFTTT - dummy"
}

func (i *dummySmsSender) Send(recipientAddress string, message string) error {
	if len([]rune(message)) > lenMessageMax {
		temp := message
		message = string([]rune(temp)[:lenMessageMax])
	}

	fmt.Printf("Sending '%s' using address '%s'\n", message, recipientAddress)

	return nil
}
