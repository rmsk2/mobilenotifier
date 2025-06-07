package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"notifier/sms"
	"notifier/tools"
)

type SmsMessage struct {
	Message string `json:"message"`
}

type SmSController struct {
	log         *log.Logger
	txt         sms.SmsSender
	addressBook sms.SmsAddressBook
}

type RecipientList struct {
	AllRecipients []string `json:"all_recipients"`
}

func NewSmsController(l *log.Logger, t sms.SmsSender, a sms.SmsAddressBook) *SmSController {
	return &SmSController{
		log:         l,
		txt:         t,
		addressBook: a,
	}
}

func (s *SmSController) Add(authWrapper tools.Wrapper) {
	http.HandleFunc("POST /notifier/api/send/{recipient}", authWrapper.Wrap(s.Handle))
	http.HandleFunc("/notifier/api/send/recipients/all", s.HandleGetAllRecipients)
}

// @Summary      Send a text message to a recipient
// @Description  Send a text message specified in the body to the recipient specified in the URL
// @Tags	     SMS
// @Param        recipient   path  string  true  "Recipient"
// @Param        message_spec  body  SmsMessage true "Specification of message to send"
// @Security     ApiKeyAuth
// @Success      200  {object} nil
// @Failure      400  {object} string
// @Failure      401  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/send/{recipient} [post]
func (s *SmSController) Handle(w http.ResponseWriter, r *http.Request) {
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

// @Summary      Retrieve all possible recipients
// @Description  Retrieve all possible recipients
// @Tags	     SMS
// @Success      200  {object} RecipientList
// @Failure      500  {object} string
// @Router       /notifier/api/send/recipients/all [get]
func (s *SmSController) HandleGetAllRecipients(w http.ResponseWriter, r *http.Request) {
	recipients, err := s.addressBook.ListRecipients()
	if err != nil {
		s.log.Printf("error reading address book: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	resp := RecipientList{
		AllRecipients: recipients,
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		s.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	s.log.Println("Created list of all possible recipients")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))
}
