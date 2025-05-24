package handler

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

type NotficationHandler struct {
	db          *repo.DBLocker
	addressBook sms.SmsAddressBook
	log         *log.Logger
	nullUuid    *tools.UUID
}

func NewNotificationHandler(l *repo.DBLocker, a sms.SmsAddressBook, lg *log.Logger) *NotficationHandler {
	nuid, _ := tools.NewUuidFromString(nullUuidStr)

	return &NotficationHandler{
		db:          l,
		addressBook: a,
		log:         lg,
		nullUuid:    nuid,
	}
}

func (n *NotficationHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
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

	raw := n.db.Lock(true)
	defer func() { n.db.Unlock(true) }()

	var notifRepo repo.NotificationRepo = repo.NewBBoltNotificationRepo(raw)
	err = notifRepo.Upsert(&notification)
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

func (n *NotficationHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	uuidRaw := r.PathValue("uuid")

	uuid, ok := tools.NewUuidFromString(uuidRaw)
	if !ok {
		n.log.Printf("Unable to parse '%s' into uuid", uuidRaw)
		http.Error(w, "UUID not wellformed", http.StatusBadRequest)
		return
	}

	raw := n.db.Lock(true)
	defer func() { n.db.Unlock(true) }()

	var notifRepo repo.NotificationRepo = repo.NewBBoltNotificationRepo(raw)
	err := notifRepo.Delete(uuid)
	if err != nil {
		n.log.Printf("error deleting notification: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	n.log.Printf("Notification with id '%s' deleted ", uuid)
}

func (n *NotficationHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	raw := n.db.Lock(false)
	defer func() { n.db.Unlock(false) }()

	var notifRepo repo.NotificationRepo = repo.NewBBoltNotificationRepo(raw)
	uuids, err := notifRepo.Filter(func(*repo.Notification) bool { return true })
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

func (n *NotficationHandler) HandleExpiry(w http.ResponseWriter, r *http.Request) {
	uuidRaw := r.PathValue("uuid")

	uuid, ok := tools.NewUuidFromString(uuidRaw)
	if !ok {
		n.log.Printf("Unable to parse '%s' into uuid", uuidRaw)
		http.Error(w, "UUID not wellformed", http.StatusBadRequest)
		return
	}

	raw := n.db.Lock(false)
	defer func() { n.db.Unlock(false) }()

	var notifRepo repo.NotificationRepo = repo.NewBBoltNotificationRepo(raw)
	notification, err := notifRepo.Get(uuid)
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
