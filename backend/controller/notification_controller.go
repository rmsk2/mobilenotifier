package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"notifier/repo"
	"notifier/sms"
	"notifier/tools"
	"time"
)

const nullUuidStr = "D20F95D6-3339-40FF-8E4E-B2F6AC439D06"

type NotificationData struct {
	WarningTime time.Time `json:"warning_time"`
	Description string    `json:"description"`
	Recipient   string    `json:"recipient"`
}

type UuidResponse struct {
	Uuid *tools.UUID `json:"uuid"`
}

type GetExpiryResponse struct {
	ExpiresAt time.Time `json:"expires_at"`
}

type ListResponse struct {
	Uuids []*tools.UUID `json:"uuids"`
}

type NotficationController struct {
	db          repo.DBSerializer
	addressBook sms.SmsAddressBook
	log         *log.Logger
	nullUuid    *tools.UUID
}

func NewNotificationController(l repo.DBSerializer, a sms.SmsAddressBook, lg *log.Logger) *NotficationController {
	nuid, _ := tools.NewUuidFromString(nullUuidStr)

	return &NotficationController{
		db:          l,
		addressBook: a,
		log:         lg,
		nullUuid:    nuid,
	}
}

func (n *NotficationController) Add() {
	http.HandleFunc("POST /notifier/api/notification", n.HandlePost)
	http.HandleFunc("/notifier/api/notification", n.HandleList)
	http.HandleFunc("DELETE /notifier/api/notification/{uuid}", n.HandleDelete)
	http.HandleFunc("/notifier/api/notification/{uuid}/expiry", n.HandleExpiry)
}

// @Summary      Create a new notification
// @Description  Create a new notification which is tracked and executed by the web service
// @Tags	     Notification
// @Accept       json
// @Param        notification_data  body  NotificationData true "Specification of notification to set"
// @Success      200  {object} UuidResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/notification [post]
func (n *NotficationController) HandlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		n.log.Println("Unable to read body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var m NotificationData
	err = json.Unmarshal(body, &m)
	if err != nil {
		n.log.Printf("Unable to parse body: '%s'", string(body))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	ok, err := n.addressBook.CheckRecipient(m.Recipient)
	if err != nil {
		n.log.Printf("error accessing recipient info: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !ok {
		t := fmt.Sprintf("recipient '%s' is unknown", m.Recipient)
		n.log.Println(t)
		http.Error(w, t, http.StatusBadRequest)
		return
	}

	var resp UuidResponse = UuidResponse{
		Uuid: tools.UUIDGen(),
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		n.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	notification := repo.Notification{
		Id:          resp.Uuid,
		Parent:      n.nullUuid,
		WarningTime: m.WarningTime,
		Description: m.Description,
		Recipient:   m.Recipient,
	}

	writeRepo, _ := n.db.Lock()
	defer func() { n.db.Unlock() }()

	err = writeRepo.Upsert(&notification)
	if err != nil {
		n.log.Printf("error creating new notification: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	n.log.Printf("Notification with id '%s' created ", resp.Uuid)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))
}

// @Summary      Delete a notification
// @Description  Delete a notfification with the specified uuid
// @Tags	     Notification
// @Param        uuid   path  string  true  "UUID of notification"
// @Success      200  {object} nil
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/notification/{uuid} [delete]
func (n *NotficationController) HandleDelete(w http.ResponseWriter, r *http.Request) {
	uuidRaw := r.PathValue("uuid")

	uuid, ok := tools.NewUuidFromString(uuidRaw)
	if !ok {
		n.log.Printf("Unable to parse '%s' into uuid", uuidRaw)
		http.Error(w, "UUID not wellformed", http.StatusBadRequest)
		return
	}

	writeRepo, _ := n.db.Lock()
	defer func() { n.db.Unlock() }()

	err := writeRepo.Delete(uuid)
	if err != nil {
		n.log.Printf("error deleting notification: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	n.log.Printf("Notification with id '%s' deleted ", uuid)
}

// @Summary      Get all existing notifications
// @Description  Get all existing notifications as a JSON list
// @Tags	     Notification
// @Accept       json
// @Success      200  {object} ListResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/notification [get]
func (n *NotficationController) HandleList(w http.ResponseWriter, r *http.Request) {
	readRepo, _ := n.db.RLock()
	defer func() { n.db.RUnlock() }()

	uuids, err := readRepo.Filter(func(*repo.Notification) bool { return true })
	if err != nil {
		n.log.Printf("error listing notification: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	resp := ListResponse{
		Uuids: uuids,
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		n.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	n.log.Println("Created list for all notifications")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))
}

// @Summary      Get expiry date of notification
// @Description  Get expiry date of notfification with the specified uuid
// @Tags	     Notification
// @Param        uuid   path  string  true  "UUID of notification"
// @Success      200  {object} nil
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/notification/{uuid}/expiry [get]
func (n *NotficationController) HandleExpiry(w http.ResponseWriter, r *http.Request) {
	uuidRaw := r.PathValue("uuid")

	uuid, ok := tools.NewUuidFromString(uuidRaw)
	if !ok {
		n.log.Printf("Unable to parse '%s' into uuid", uuidRaw)
		http.Error(w, "UUID not wellformed", http.StatusBadRequest)
		return
	}

	readRepo, _ := n.db.RLock()
	defer func() { n.db.RUnlock() }()

	notification, err := readRepo.Get(uuid)
	if err != nil {
		n.log.Printf("error retrieving notification: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if notification == nil {
		n.log.Printf("requested notification id '%s' was not found", uuid)
		http.Error(w, "notification not found", http.StatusBadRequest)
		return
	}

	resp := GetExpiryResponse{
		ExpiresAt: notification.WarningTime,
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		n.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	n.log.Printf("Returned expiry data for '%s'", uuid)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))
}
