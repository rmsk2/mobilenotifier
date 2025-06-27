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
	recipientMap map[string]Recipient
}

func NewMailNotifier(h string, p uint16, s string, pw string, mailMartin string, mailPush string) *mailNotifier {
	res := &mailNotifier{
		Host:         h,
		Port:         p,
		SenderAdress: s,
		Password:     pw,
		Subject:      "Benachrichtigung",
		recipientMap: makeRecipientMap(),
	}

	help := res.recipientMap[idMartin]
	help.Address = mailMartin
	res.recipientMap[idMartin] = help

	help = res.recipientMap[idPush]
	help.Address = mailPush
	res.recipientMap[idPush] = help

	return res
}

func (m *mailNotifier) SetSubject(s string) {
	m.Subject = s
}

func (m *mailNotifier) CheckRecipient(id string) (bool, error) {
	_, ok := m.recipientMap[id]
	return ok, nil
}

func (m *mailNotifier) ListRecipients() ([]RecipientInfo, error) {
	return listRecipientsOnMap(m.recipientMap), nil
}

func (m *mailNotifier) Send(recipientId string, message string) error {
	v, ok := m.recipientMap[recipientId]
	if !ok {
		return fmt.Errorf("recipientid %s is unknown", recipientId)
	}

	to := []string{v.Address}

	msg := fmt.Sprintf("From: %s\r\n", m.SenderAdress)
	msg += fmt.Sprintf("To: %s\r\n", v.Address)
	msg += fmt.Sprintf("Subject: %s\r\n", m.Subject)
	msg += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	msg += "\r\n"
	msg += message

	auth := smtp.PlainAuth("", m.SenderAdress, m.Password, m.Host)
	return smtp.SendMail(fmt.Sprintf("%s:%d", m.Host, m.Port), auth, m.SenderAdress, to, []byte(msg))
}
