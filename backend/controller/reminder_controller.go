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
	"sort"
	"strconv"
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
	NextEvent   time.Time         `json:"next_occurrance"`
}

type ExtReminder struct {
	Reminder  *repo.Reminder `json:"reminder"`
	NextEvent time.Time      `json:"next_occurrance"`
}

type ReminderListResponse struct {
	Reminders []*ExtReminder `json:"reminders"`
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
	http.HandleFunc("/notifier/api/reminder/views/bymonth", n.HandleViewByMonth)
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
		n.log.Printf("Unable to parse body '%s'. Error: %v", string(body), err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if len(m.Recipients) == 0 {
		n.log.Printf("Illegal number of recipients: %d", len(m.Recipients))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if len(m.WarningAt) == 0 {
		n.log.Printf("Illegal number of warning types: %d", len(m.WarningAt))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	for _, j := range m.WarningAt {
		if (j < repo.MorningBefore) || (j > repo.SameDay) {
			n.log.Printf("Illegal warning type: %d", j)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}

	if (repo.WarningType(m.Kind) < repo.WarningType(repo.Anniversary)) || (m.Kind > repo.OneShot) {
		n.log.Printf("Illegal kind of reminder type: %d", m.Kind)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if m.Description == "" {
		n.log.Printf("Description is empty. This makes no sense")
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

// @Summary      Get all existing reminders
// @Description  Get all existing reminders as a JSON list. This list is sorted in ascending order with respect to next_occurance
// @Tags	     Reminder
// @Success      200  {object} ReminderListResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/reminder [get]
func (n *ReminderController) HandleList(w http.ResponseWriter, r *http.Request) {
	n.HandleFiltered(w, r, func(*repo.Reminder) bool { return true }, time.Now())
}

// @Summary      Get all existing reminders for given month and year
// @Description  Get all existing reminders for given month and year as a JSON list. This list is sorted in ascending order with respect to next_occurance
// @Tags	     Reminder
// @Param        month    query     int  true  "month to look at"
// @Param        year    query     int  true  "year to look at"
// @Success      200  {object} ReminderListResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/reminder/views/bymonth [get]
func (n *ReminderController) HandleViewByMonth(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("month") == "" {
		n.log.Printf("Query parameter for month not found")
		http.Error(w, "Illegal month", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("year") == "" {
		n.log.Printf("Query parameter for year not found")
		http.Error(w, "Illegal year", http.StatusBadRequest)
		return
	}

	month, err := strconv.Atoi(r.URL.Query().Get("month"))
	if err != nil {
		n.log.Printf("error parsing month parameter: %v", err)
		http.Error(w, "Illegal month", http.StatusBadRequest)
		return
	}

	if (month < 1) || (month > 12) {
		n.log.Printf("illegal month parameter: %d", month)
		http.Error(w, "Illegal month", http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if err != nil {
		n.log.Printf("error parsing year parameter: %v", err)
		http.Error(w, "Illegal year", http.StatusBadRequest)
		return
	}

	if year < 0 {
		n.log.Printf("illegal year parameter: %d", month)
		http.Error(w, "Illegal year", http.StatusBadRequest)
		return
	}

	refTimeStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	oneMillisecond := time.Millisecond
	refTimeStart = refTimeStart.Add(-oneMillisecond)

	help := month - 1
	help++
	help = help % 12
	help++
	refTimeEnd := time.Date(year, time.Month(help), 1, 0, 0, 0, 0, time.UTC)

	timeFilter := func(r *repo.Reminder) bool {
		t := logic.RefTimeMap[r.Kind](r, refTimeStart)
		return (t.Compare(refTimeStart) != -1) && (t.Compare(refTimeEnd) != 1)
	}

	n.HandleFiltered(w, r, timeFilter, refTimeStart)
}

func (n *ReminderController) HandleFiltered(w http.ResponseWriter, r *http.Request, filterFunc repo.ReminderPredicate, refNow time.Time) {
	_, readRepo := n.db.RLock()
	defer func() { n.db.RUnlock() }()

	allReminders, err := readRepo.Filter(filterFunc)
	if err != nil {
		n.log.Printf("error listing notification: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	res := []*ExtReminder{}
	for _, j := range allReminders {
		i := &ExtReminder{
			Reminder:  j,
			NextEvent: logic.RefTimeMap[j.Kind](j, refNow),
		}
		res = append(res, i)
	}

	sort.SliceStable(res, func(i, j int) bool {
		return res[i].NextEvent.Compare(res[j].NextEvent) == -1
	})

	resp := ReminderListResponse{
		Reminders: res,
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
// @Description  Get basic information for existing reminders as a JSON list. The entries are sorted in ascending order with respect to next_occurrance.
// @Tags	     Reminder
// @Param        max_entries    query     int  true  "maximum number of entries to return"
// @Success      200  {object} OverviewResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/reminder/views/basic [get]
func (n *ReminderController) HandleOverview(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("max_entries") == "" {
		n.log.Printf("Query parameter for maximum number of entries not found")
		http.Error(w, "Illegal month", http.StatusBadRequest)
		return
	}

	maxEntries, err := strconv.Atoi(r.URL.Query().Get("max_entries"))
	if err != nil {
		n.log.Printf("error parsing max_entries parameter: %v", err)
		http.Error(w, "Illegal max_entries parameter", http.StatusBadRequest)
		return
	}

	_, readRepo := n.db.RLock()
	defer func() { n.db.RUnlock() }()

	refTime := time.Now()

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
			NextEvent:   logic.RefTimeMap[j.Kind](j, refTime),
		}
		responses = append(responses, &o)
	}

	sort.SliceStable(responses, func(i, j int) bool {
		return responses[i].NextEvent.Compare(responses[j].NextEvent) == -1
	})

	var limit int

	if (maxEntries <= 0) || (maxEntries > len(responses)) {
		limit = len(responses)
	} else {
		limit = maxEntries
	}

	resp := OverviewResponse{
		Reminders: responses[:limit],
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
