package sms

import (
	"fmt"
	"net/smtp"
)

type mailNotifier struct {
	Host         string
	Port         uint16
	SenderAdress string
	Password     string
	Subject      string
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
const addrMail = "recipient"

func AddMailRecipients(a *AddressBook) {
	martin := Recipient{
		DisplayName: displayMartinMail,
		Id:          idMartinMail,
		Address:     addrMail,
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
