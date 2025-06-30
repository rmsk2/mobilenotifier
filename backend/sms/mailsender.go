package sms

import (
	"fmt"
	"net/smtp"
	"os"
	"strconv"
)

type mailNotifier struct {
	Host         string
	Port         uint16
	SenderAdress string
	Password     string
	Subject      string
}

const envMailServer = "MN_MAIL_SERVER"
const envServerPort = "MN_MAIL_SERVER_PORT"
const envSenderAddress = "MN_MAIL_SENDER_ADDR"
const envServerPassword = "MN_MAIL_SENDER_PW"
const envMartinMail = "MN_MAIL_RECIPIENT"

var mailMartin string = "niemand@nix.de"

func NewMailNotifierFromEnvironment() (*mailNotifier, error) {
	mailServer, ok := os.LookupEnv(envMailServer)
	if !ok {
		return nil, fmt.Errorf("no mailer config found")
	}

	mailPort, ok := os.LookupEnv(envServerPort)
	if !ok {
		return nil, fmt.Errorf("no mailer config found")
	}

	port, err := strconv.ParseUint(mailPort, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("no mailer config found")
	}

	port16 := (uint16)(port)

	senderAddr, ok := os.LookupEnv(envSenderAddress)
	if !ok {
		return nil, fmt.Errorf("no mailer config found")
	}

	password, ok := os.LookupEnv(envServerPassword)
	if !ok {
		return nil, fmt.Errorf("no mailer config found")
	}

	mailMartin, ok = os.LookupEnv(envMartinMail)
	if !ok {
		return nil, fmt.Errorf("no mailer config found")
	}

	return NewMailNotifier(mailServer, port16, senderAddr, password), nil
}

func NewMailNotifier(h string, p uint16, s string, pw string) *mailNotifier {
	res := &mailNotifier{
		Host:         h,
		Port:         p,
		SenderAdress: s,
		Password:     pw,
		Subject:      "Benachrichtigung",
	}

	return res
}

const displayMartinMail = "Martin via Mail"
const idMartinMail = "0E69B617-12D0-4491-ADD8-D103CF3925A1"

func AddMailRecipients(a *AddressBook) {
	martin := Recipient{
		DisplayName: displayMartinMail,
		Id:          idMartinMail,
		Address:     mailMartin,
		AddrType:    TypeMail,
	}

	a.AddRecipient(martin)
}

func (m *mailNotifier) SetSubject(s string) {
	m.Subject = s
}

func (m *mailNotifier) Send(recipientAddress string, message string) error {
	to := []string{recipientAddress}

	msg := fmt.Sprintf("From: %s\r\n", m.SenderAdress)
	msg += fmt.Sprintf("To: %s\r\n", recipientAddress)
	msg += fmt.Sprintf("Subject: %s\r\n", m.Subject)
	msg += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	msg += "\r\n"
	msg += message

	auth := smtp.PlainAuth("", m.SenderAdress, m.Password, m.Host)
	return smtp.SendMail(fmt.Sprintf("%s:%d", m.Host, m.Port), auth, m.SenderAdress, to, []byte(msg))
}
