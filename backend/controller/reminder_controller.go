package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"notifier/logic"
	"notifier/repo"
	"notifier/sms"
	"notifier/tools"
	"time"
)

type ReminderData struct {
	Kind        repo.ReminderType  `json:"kind"`
	Param       int                `json:"param"`
	WarningAt   []repo.WarningType `json:"warning_at"`
	Spec        time.Time          `json:"warning_time"`
	Description string             `json:"description"`
	Recipients  []string           `json:"recipients"`
}

type ReminderResult struct {
	ReminderData
	Id *tools.UUID `json:"id"`
}

type ReminderResponse struct {
	Found bool           `json:"found"`
	Data  *repo.Reminder `json:"data"`
}

type ReminderController struct {
	db          repo.DBSerializer
	addressBook sms.SmsAddressBook
	log         *log.Logger
}

type ReminderOverview struct {
	Id          *tools.UUID       `json:"id"`
	Description string            `json:"description"`
	Kind        repo.ReminderType `json:"kind"`
}

type OverviewResponse struct {
	Reminders []*ReminderOverview `json:"reminders"`
}

func NewReminderController(l repo.DBSerializer, a sms.SmsAddressBook, lg *log.Logger) *ReminderController {
	return &ReminderController{
		db:          l,
		addressBook: a,
		log:         lg,
	}
}

func (n *ReminderController) Add() {
	http.HandleFunc("POST /notifier/api/reminder", n.HandlePost)
	http.HandleFunc("/notifier/api/reminder", n.HandleList)
	http.HandleFunc("/notifier/api/reminder/views/basic", n.HandleOverview)
	http.HandleFunc("PUT /notifier/api/reminder/{uuid}", n.HandlePostUpsert)
	http.HandleFunc("DELETE /notifier/api/reminder/{uuid}", n.HandleDelete)
	http.HandleFunc("/notifier/api/reminder/{uuid}", n.HandleGet)
}

// @Summary      Create a new reminder
// @Description  Create a new reminder which is tracked and executed by the web service
// @Tags	     Reminder
// @Accept       json
// @Param        reminder_data  body  ReminderData true "Specification of reminder to set"
// @Success      200  {object} UuidResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/reminder [post]
func (n *ReminderController) HandlePost(w http.ResponseWriter, r *http.Request) {
	n.HandleUpsert(w, r, tools.UUIDGen())
}

// @Summary      Modify or create a reminder
// @Description  Create a new or modify an existing reminder with the id specified in the path. This also regenerates all notifications currently associated with the reminder.
// @Tags	     Reminder
// @Accept       json
// @Param        uuid   path  string  true  "UUID of reminder"
// @Param        reminder_data  body  ReminderData true "Specification of reminder to set"
// @Success      200  {object} UuidResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/reminder/{uuid} [put]
func (n *ReminderController) HandlePostUpsert(w http.ResponseWriter, r *http.Request) {
	uuidRaw := r.PathValue("uuid")

	uuid, ok := tools.NewUuidFromString(uuidRaw)
	if !ok {
		n.log.Printf("Unable to parse '%s' into uuid", uuidRaw)
		http.Error(w, "UUID not wellformed", http.StatusBadRequest)
		return
	}

	n.HandleUpsert(w, r, uuid)
}

func (n *ReminderController) HandleUpsert(w http.ResponseWriter, r *http.Request, uuid *tools.UUID) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		n.log.Println("Unable to read body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var m ReminderData
	err = json.Unmarshal(body, &m)
	if err != nil {
		n.log.Printf("Unable to parse body: '%s'", string(body))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	for _, j := range m.Recipients {
		ok, err := n.addressBook.CheckRecipient(j)
		if err != nil {
			n.log.Printf("error accessing recipient info: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		if !ok {
			t := fmt.Sprintf("recipient '%s' is unknown", j)
			n.log.Println(t)
			http.Error(w, t, http.StatusBadRequest)
			return
		}
	}

	var resp UuidResponse = UuidResponse{
		Uuid: uuid,
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		n.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	reminder := repo.Reminder{
		Id:          resp.Uuid,
		Kind:        m.Kind,
		Param:       m.Param,
		WarningAt:   m.WarningAt,
		Spec:        m.Spec,
		Description: m.Description,
		Recipients:  m.Recipients,
	}

	nWriteRepo, writeRepo := n.db.Lock()
	defer func() { n.db.Unlock() }()

	err = repo.ClearNotifications(nWriteRepo, reminder.Id)
	if err != nil {
		n.log.Printf("error clearing possibly existing notifications: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	err = writeRepo.Upsert(&reminder)
	if err != nil {
		n.log.Printf("error creating/updating reminder: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// ToDo: Attempt to cleanup DB if this fails
	err = logic.ProcessNewUuid(nWriteRepo, writeRepo, &reminder)
	if err != nil {
		n.log.Printf("error creating notifications for new/updated reminder: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	n.log.Printf("Reminder with id '%s' created ", resp.Uuid)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))

}

// @Summary      Delete a reminder
// @Description  Delete a reminder with the specified uuid and all notifications associated with it
// @Tags	     Reminder
// @Param        uuid   path  string  true  "UUID of reminder"
// @Success      200  {object} nil
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/reminder/{uuid} [delete]
func (n *ReminderController) HandleDelete(w http.ResponseWriter, r *http.Request) {
	uuidRaw := r.PathValue("uuid")

	uuid, ok := tools.NewUuidFromString(uuidRaw)
	if !ok {
		n.log.Printf("Unable to parse '%s' into uuid", uuidRaw)
		http.Error(w, "UUID not wellformed", http.StatusBadRequest)
		return
	}

	nWriteRepo, writeRepo := n.db.Lock()
	defer func() { n.db.Unlock() }()

	err := writeRepo.Delete(uuid)
	if err != nil {
		n.log.Printf("error deleting reminder: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	err = repo.ClearNotifications(nWriteRepo, uuid)
	if err != nil {
		n.log.Printf("error deleting notifications for reminder '%s': %v", uuid, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	n.log.Printf("reminder with id '%s' deleted ", uuid)
}

// @Summary      Get a reminder
// @Description  Get a reminder with the specified uuid
// @Tags	     Reminder
// @Param        uuid   path  string  true  "UUID of reminder"
// @Success      200  {object} ReminderResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/reminder/{uuid} [get]
func (n *ReminderController) HandleGet(w http.ResponseWriter, r *http.Request) {
	uuidRaw := r.PathValue("uuid")

	uuid, ok := tools.NewUuidFromString(uuidRaw)
	if !ok {
		n.log.Printf("Unable to parse '%s' into uuid", uuidRaw)
		http.Error(w, "UUID not wellformed", http.StatusBadRequest)
		return
	}

	_, readRepo := n.db.RLock()
	defer func() { n.db.RUnlock() }()

	reminder, err := readRepo.Get(uuid)
	if err != nil {
		n.log.Printf("error getting reminder: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var response ReminderResponse

	if reminder == nil {
		response = ReminderResponse{
			Found: false,
			Data:  nil,
		}
	} else {
		response = ReminderResponse{
			Found: true,
			Data:  reminder,
		}
	}

	data, err := json.Marshal(&response)
	if err != nil {
		n.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))

	n.log.Printf("reminder with id '%s' read from repo ", uuid)
}

type ReminderListResponse struct {
	Reminders []*repo.Reminder `json:"reminders"`
}

// @Summary      Get all existing reminders
// @Description  Get all existing reminders as a JSON list
// @Tags	     Reminder
// @Success      200  {object} ReminderListResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/reminder [get]
func (n *ReminderController) HandleList(w http.ResponseWriter, r *http.Request) {
	_, readRepo := n.db.RLock()
	defer func() { n.db.RUnlock() }()

	allReminders, err := readRepo.Filter(func(*repo.Reminder) bool { return true })
	if err != nil {
		n.log.Printf("error listing notification: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	resp := ReminderListResponse{
		Reminders: allReminders,
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		n.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	n.log.Println("Created list for all reminders")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))
}

// @Summary      Get basic information about existing reminders
// @Description  Get basic information for existing reminders as a JSON list
// @Tags	     Reminder
// @Success      200  {object} OverviewResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/reminder/views/basic [get]
func (n *ReminderController) HandleOverview(w http.ResponseWriter, r *http.Request) {
	_, readRepo := n.db.RLock()
	defer func() { n.db.RUnlock() }()

	allReminders, err := readRepo.Filter(func(*repo.Reminder) bool { return true })
	if err != nil {
		n.log.Printf("error listing reminders: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	responses := []*ReminderOverview{}
	for _, j := range allReminders {
		o := ReminderOverview{
			Id:          j.Id,
			Description: j.Description,
			Kind:        j.Kind,
		}
		responses = append(responses, &o)
	}

	resp := OverviewResponse{
		Reminders: responses,
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		n.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	n.log.Println("Created overview for all reminders")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))
}
