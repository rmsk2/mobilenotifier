package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"notifier/sms"
)

type SmsMessage struct {
	Message string `json:"message"`
}

type SmSHandler struct {
	log         *log.Logger
	txt         sms.SmsSender
	addressBook sms.SmsAddressBook
}

func NewSmsHandler(l *log.Logger, t sms.SmsSender, a sms.SmsAddressBook) *SmSHandler {
	return &SmSHandler{
		log:         l,
		txt:         t,
		addressBook: a,
	}
}

func (s *SmSHandler) Handle(w http.ResponseWriter, r *http.Request) {
	recipient := r.PathValue("recipient")
	ok, err := s.addressBook.CheckRecipient(recipient)

	if err != nil {
		s.log.Printf("error accessing recipient info: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !ok {
		t := fmt.Sprintf("recipient '%s' is unknown", recipient)
		s.log.Println(t)
		http.Error(w, t, http.StatusBadRequest)
		return
	}

	s.log.Printf("Sending SMS to '%s'", recipient)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.log.Println("Unable to read body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var m SmsMessage
	err = json.Unmarshal(body, &m)
	if err != nil {
		s.log.Printf("Unable to parse body: '%s'", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.txt.Send(recipient, m.Message)
	if err != nil {
		s.log.Printf("Sending SMS failed: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	s.log.Printf("SMS with message '%s' successfully sent", m.Message)
}
