package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const lenMessageMax = 160
const displayMartin = "Martin"
const displayPush = "Pushmessage"
const idMartin = "0D69B617-12D0-4491-ADD8-D103CF3925A1"
const idPush = "F55C84F3-A2C7-46DD-AF06-27AFF7FCCC16"
const addrSMS = "SendSMS1"
const addrPush = "SendPush1"

type Recipient struct {
	DisplayName string `json:"display_name"`
	Id          string `json:"id"`
	Address     string `json:"address"`
}

type RecipientInfo struct {
	Id          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type SmsAddressBook interface {
	CheckRecipient(r string) (bool, error)
	ListRecipients() ([]RecipientInfo, error)
}

type SmsSender interface {
	Send(recipient string, message string) error
}

func NewIftttSender(apiKey string) *iftttSmsSender {
	res := new(iftttSmsSender)

	martin := Recipient{
		DisplayName: displayMartin,
		Id:          idMartin,
		Address:     addrSMS,
	}

	push := Recipient{
		DisplayName: displayPush,
		Id:          idPush,
		Address:     addrPush,
	}

	res.recipientMap = map[string]Recipient{
		idMartin: martin,
		idPush:   push,
	}

	res.apiKey = apiKey

	return res
}

type ifftBody struct {
	Value1 string `json:"value1"`
}

type iftttSmsSender struct {
	recipientMap map[string]Recipient
	apiKey       string
}

func (i *iftttSmsSender) CheckRecipient(id string) (bool, error) {
	_, ok := i.recipientMap[id]
	return ok, nil
}

func (i *iftttSmsSender) ListRecipients() ([]RecipientInfo, error) {
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

func (i *iftttSmsSender) Send(resipientId string, message string) error {
	v, ok := i.recipientMap[resipientId]
	if !ok {
		return fmt.Errorf("recipientid %s is unknown", resipientId)
	}

	requestURL := fmt.Sprintf("https://maker.ifttt.com/trigger/%s/with/key/%s", v.Address, i.apiKey)

	if len([]rune(message)) > lenMessageMax {
		temp := message
		message = string([]rune(temp)[:lenMessageMax])
	}

	var ifftData ifftBody
	ifftData.Value1 = message

	body, err := json.Marshal(&ifftData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if (res.StatusCode < 200) || (res.StatusCode >= 300) {
		return fmt.Errorf("server responded with error code %d", res.StatusCode)
	}

	// Ignore body
	io.ReadAll(res.Body)

	return nil
}
