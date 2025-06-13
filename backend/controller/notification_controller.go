package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"notifier/repo"
	"notifier/sms"
	"notifier/tools"
	"time"
)

type GetResponse struct {
	Found bool               `json:"found"`
	Data  *repo.Notification `json:"data"`
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
}

func NewNotificationController(l repo.DBSerializer, a sms.SmsAddressBook, lg *log.Logger) *NotficationController {
	return &NotficationController{
		db:          l,
		addressBook: a,
		log:         lg,
	}
}

func (n *NotficationController) Add() {
	http.HandleFunc("GET /notifier/api/notification", n.HandleList)
	http.HandleFunc("DELETE /notifier/api/notification/{uuid}", n.HandleDelete)
	http.HandleFunc("GET /notifier/api/notification/siblings/{uuid}", n.HandleGetSiblings)
	http.HandleFunc("GET /notifier/api/notification/{uuid}", n.HandleGet)
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

// @Summary      Get a notification
// @Description  Get a notfification with the specified uuid
// @Tags	     Notification
// @Param        uuid   path  string  true  "UUID of notification"
// @Success      200  {object} GetResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/notification/{uuid} [get]
func (n *NotficationController) HandleGet(w http.ResponseWriter, r *http.Request) {
	uuidRaw := r.PathValue("uuid")

	uuid, ok := tools.NewUuidFromString(uuidRaw)
	if !ok {
		n.log.Printf("Unable to parse '%s' into uuid", uuidRaw)
		http.Error(w, "UUID not wellformed", http.StatusBadRequest)
		return
	}

	readRepo, _ := n.db.RLock()
	defer func() { n.db.RUnlock() }()

	notificationData, err := readRepo.Get(uuid)
	if err != nil {
		n.log.Printf("error reading notification: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var resp GetResponse

	if notificationData == nil {
		resp = GetResponse{
			Found: false,
			Data:  nil,
		}
	} else {
		resp = GetResponse{
			Found: true,
			Data:  notificationData,
		}
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		n.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))

	n.log.Printf("Notification with id '%s' read ", uuid)
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
	n.HandleFilter(w, r, func(*repo.Notification) bool {
		return true
	})
}

// @Summary      Get the ids of all notifications belonging to a reminder
// @Description  Get the ids of all notifications belonging to a reminder with the specified uuid
// @Tags	     Notification
// @Param        uuid   path  string  true  "UUID of notification"
// @Success      200  {object} ListResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/notification/siblings/{uuid} [get]
func (n *NotficationController) HandleGetSiblings(w http.ResponseWriter, r *http.Request) {
	uuidRaw := r.PathValue("uuid")

	uuid, ok := tools.NewUuidFromString(uuidRaw)
	if !ok {
		n.log.Printf("Unable to parse '%s' into uuid", uuidRaw)
		http.Error(w, "UUID not wellformed", http.StatusBadRequest)
		return
	}

	n.HandleFilter(w, r, func(notif *repo.Notification) bool {
		return notif.Parent.IsEqual(uuid)
	})
}

func (n *NotficationController) HandleFilter(w http.ResponseWriter, r *http.Request, filterFunc repo.NotificationPredicate) {
	readRepo, _ := n.db.RLock()
	defer func() { n.db.RUnlock() }()

	uuids, err := readRepo.Filter(filterFunc)
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

	n.log.Println("Created filtered list of all notifications")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))
}
