package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SmsAddressBook interface {
	CheckRecipient(r string) (bool, error)
	ListRecipients() ([]string, error)
}

type SmsSender interface {
	Send(recipient string, message string) error
}

func NewIftttSender(apiKey string) *iftttSmsSender {
	res := new(iftttSmsSender)
	res.recipientMap = map[string]string{
		"martin": "SendSMS1",
	}

	res.apiKey = apiKey

	return res
}

type ifftBody struct {
	Value1 string `json:"value1"`
}

type iftttSmsSender struct {
	recipientMap map[string]string
	apiKey       string
}

func (i *iftttSmsSender) CheckRecipient(r string) (bool, error) {
	_, ok := i.recipientMap[r]
	return ok, nil
}

func (i *iftttSmsSender) ListRecipients() ([]string, error) {
	keys := make([]string, len(i.recipientMap))

	for k := range i.recipientMap {
		keys = append(keys, k)
	}

	return keys, nil
}

func (i *iftttSmsSender) Send(recipient string, message string) error {
	v, ok := i.recipientMap[recipient]
	if !ok {
		return fmt.Errorf("recipient %s is unknown", recipient)
	}

	requestURL := fmt.Sprintf("https://maker.ifttt.com/trigger/%s/with/key/%s", v, i.apiKey)

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
