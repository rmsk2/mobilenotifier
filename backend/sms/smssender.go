package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const lenMessageMax = 160

type SmsSender interface {
	Send(recipientAddress string, message string) error
}

func NewIftttSender(apiKey string) *iftttSmsSender {
	res := new(iftttSmsSender)
	res.apiKey = apiKey

	return res
}

type ifftBody struct {
	Value1 string `json:"value1"`
}

type iftttSmsSender struct {
	apiKey string
}

func (i *iftttSmsSender) Send(recipientAddress string, message string) error {
	requestURL := fmt.Sprintf("https://maker.ifttt.com/trigger/%s/with/key/%s", recipientAddress, i.apiKey)

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
	defer res.Body.Close()

	// Ignore body
	io.ReadAll(res.Body)

	if (res.StatusCode < 200) || (res.StatusCode >= 300) {
		return fmt.Errorf("server responded with error code %d", res.StatusCode)
	}

	return nil
}
