package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"notifier/tools"
	"os"
)

const envLocalSenderAddress = "MN_LOCAL_SENDER_URL"
const envLocalSenderToken = "MN_LOCAL_SENDER_TOKEN"

type LocalSmsSender struct {
	ServiceUrl string
	Jwt        string
	Client     *http.Client
}

type SendRequest struct {
	Message string `json:"message"`
	PhoneNr string `json:"phone_nr"`
}

func NewLocalSender(url string, jwt string, c *http.Client) *LocalSmsSender {
	res := LocalSmsSender{
		ServiceUrl: url,
		Jwt:        jwt,
		Client:     c,
	}

	return &res
}

func NewLocalSenderFromEnvironment() (*LocalSmsSender, error) {
	url, ok := os.LookupEnv(envLocalSenderAddress)
	if !ok {
		return nil, fmt.Errorf("Environment variable %s not set", envLocalSenderAddress)
	}

	token, ok := os.LookupEnv(envLocalSenderToken)
	if !ok {
		return nil, fmt.Errorf("Environment variable %s not set", envLocalSenderToken)
	}

	client, err := tools.MakeCustomHttpClient()
	if err != nil {
		return nil, fmt.Errorf("Unable to create custom HTTP client object: %v", err)
	}

	return NewLocalSender(url, token, client), nil
}

func (l *LocalSmsSender) GetName() string {
	return "local"
}

func (l *LocalSmsSender) Send(recipientAddress string, message string) error {
	if len([]rune(message)) > lenMessageMax {
		temp := message
		message = string([]rune(temp)[:lenMessageMax])
	}

	smsReq := SendRequest{
		Message: message,
		PhoneNr: recipientAddress,
	}

	body, err := json.Marshal(&smsReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, l.ServiceUrl, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token", l.Jwt)

	res, err := l.Client.Do(req)
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
