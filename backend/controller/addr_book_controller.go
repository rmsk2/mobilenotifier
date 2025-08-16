package controller

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"notifier/logic"
	"notifier/repo"
	"notifier/tools"
	"sort"
	"strings"
)

type RecipientResponse GetResponseGeneric[*repo.Recipient]

type RecipientData struct {
	DisplayName string `json:"display_name"`
	Address     string `json:"address"`
	AddrType    string `json:"addr_type"`
	IsDefault   bool   `json:"is_default"`
}

type AddressBookController struct {
	db         repo.DBSerializer
	dbRemNotif repo.DBSerializer
	log        *log.Logger
	genRead    func(repo.DbType) repo.AddrBookRead
	genWrite   func(repo.DbType) repo.AddrBookWrite
}

func NewAddressBookController(l repo.DBSerializer, lRemNotif repo.DBSerializer, lg *log.Logger, g func(repo.DbType) *repo.BBoltAddrBookRepo) *AddressBookController {
	genR := func(db repo.DbType) repo.AddrBookRead {
		return g(db)
	}

	genW := func(db repo.DbType) repo.AddrBookWrite {
		return g(db)
	}

	return &AddressBookController{
		db:         l,
		dbRemNotif: lRemNotif,
		log:        lg,
		genRead:    genR,
		genWrite:   genW,
	}
}

func (a *AddressBookController) AddHandlers() {
	http.HandleFunc("GET /notifier/api/addressbook", a.HandleList)
	http.HandleFunc("POST /notifier/api/addressbook", a.HandleCreate)
	http.HandleFunc("DELETE /notifier/api/addressbook/{uuid}", a.HandleDelete)
	http.HandleFunc("GET /notifier/api/addressbook/{uuid}", a.HandleGet)
	http.HandleFunc("PUT /notifier/api/addressbook/{uuid}", a.HandleUpsert)
}

// @Summary      Delete an address book entry
// @Description  Delete an address book entry with the specified uuid
// @Tags	     AddressBook
// @Param        uuid   path  string  true  "UUID of entry"
// @Success      200  {object} nil
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/addressbook/{uuid} [delete]
func (a *AddressBookController) HandleDelete(w http.ResponseWriter, r *http.Request) {
	uuidRaw := r.PathValue("uuid")

	uuid, ok := tools.NewUuidFromString(uuidRaw)
	if !ok {
		a.log.Printf("Unable to parse '%s' into uuid", uuidRaw)
		http.Error(w, "UUID not wellformed", http.StatusBadRequest)
		return
	}

	// Obatain lock on Reminders and Notifications first. If we consistently
	// do this we can prevent deadlocks
	nWrite, remWrite := a.dbRemNotif.Lock()
	defer a.dbRemNotif.Unlock()

	repoWrite := repo.LockAndGetRepoRW(a.db, a.genWrite)
	defer func() { a.db.Unlock() }()

	err := logic.DeleteAddrBookEntry(nWrite, remWrite, repoWrite, uuid)
	if err != nil {
		a.log.Printf("error deleting from database: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	a.log.Printf("address book entry with id %s deleted", uuid)
}

// @Summary      Get all existing address book entries
// @Description  Get all existing address book entries as a JSON list
// @Tags	     AddressBook
// @Accept       json
// @Success      200  {object} []repo.Recipient
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/addressbook [get]
func (a *AddressBookController) HandleList(w http.ResponseWriter, r *http.Request) {
	readRepo := repo.LockAndGetRepoR(a.db, a.genRead)
	defer func() { a.db.RUnlock() }()

	allEntries, err := readRepo.Filter(func(*repo.Recipient) bool { return true })
	if err != nil {
		a.log.Printf("error listing address book entries: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	responses := []*repo.Recipient{}
	responses = append(responses, allEntries...)

	sort.SliceStable(responses, func(i, j int) bool {
		return strings.Compare(responses[i].DisplayName, responses[j].DisplayName) == -1
	})

	data, err := json.Marshal(&responses)
	if err != nil {
		a.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	a.log.Println("Created listing of all address book entries")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))
}

// @Summary      Get an address book entry
// @Description  Get an address book entry with the specified uuid
// @Tags	     AddressBook
// @Param        uuid   path  string  true  "UUID of address book entry"
// @Success      200  {object} RecipientResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/addressbook/{uuid} [get]
func (a *AddressBookController) HandleGet(w http.ResponseWriter, r *http.Request) {
	uuidRaw := r.PathValue("uuid")

	uuid, ok := tools.NewUuidFromString(uuidRaw)
	if !ok {
		a.log.Printf("Unable to parse '%s' into uuid", uuidRaw)
		http.Error(w, "UUID not wellformed", http.StatusBadRequest)
		return
	}

	repoRead := repo.LockAndGetRepoR(a.db, a.genRead)
	defer func() { a.db.RUnlock() }()

	recipient, err := repoRead.Get(uuid)
	if err != nil {
		a.log.Printf("error reading from database: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var response RecipientResponse

	if recipient == nil {
		response = RecipientResponse{
			Found: false,
			Data:  nil,
		}
	} else {
		response = RecipientResponse{
			Found: true,
			Data:  recipient,
		}
	}

	data, err := json.Marshal(&response)
	if err != nil {
		a.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))

	a.log.Printf("address book entry with id '%s' read from repo ", uuid)
}

// @Summary      Modify or create an address book entry
// @Description  Create a new or modify an existing an address book entry with the id specified in the path.
// @Tags	     AddressBook
// @Accept       json
// @Param        uuid   path  string  true  "UUID of address book entry"
// @Param        addr_book_data  body  RecipientData true "Specification of address book entry to set"
// @Success      200  {object} UuidResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/addressbook/{uuid} [put]
func (a *AddressBookController) HandleUpsert(w http.ResponseWriter, r *http.Request) {
	uuidRaw := r.PathValue("uuid")

	uuid, ok := tools.NewUuidFromString(uuidRaw)
	if !ok {
		a.log.Printf("Unable to parse '%s' into uuid", uuidRaw)
		http.Error(w, "UUID not wellformed", http.StatusBadRequest)
		return
	}

	a.HandleUpsertRaw(w, r, uuid)
}

func (a *AddressBookController) HandleUpsertRaw(w http.ResponseWriter, r *http.Request, uuid *tools.UUID) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.log.Println("Unable to read body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var m RecipientData
	err = json.Unmarshal(body, &m)
	if err != nil {
		a.log.Printf("Unable to parse body '%s'. Error: %v", string(body), err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if (m.AddrType == "") || (m.Address == "") || (m.DisplayName == "") {
		a.log.Printf("Incorrect contents in body '%s'", string(body))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	repoWrite := repo.LockAndGetRepoRW(a.db, a.genWrite)
	defer func() { a.db.Unlock() }()

	recipient := repo.Recipient{
		Id:          uuid,
		DisplayName: m.DisplayName,
		Address:     m.Address,
		AddrType:    m.AddrType,
		IsDefault:   m.IsDefault,
	}

	err = repoWrite.Upsert(&recipient)
	if err != nil {
		a.log.Printf("error writing to db '%v'", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var resp UuidResponse = UuidResponse{
		Uuid: uuid,
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		a.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	a.log.Printf("Address book entry with id '%s' created ", resp.Uuid)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))
}

// @Summary      Create an address book entry
// @Description  Create a new address book entry
// @Tags	     AddressBook
// @Accept       json
// @Param        addr_book_data  body  RecipientData true "Specification of address book entry to set"
// @Success      200  {object} UuidResponse
// @Failure      400  {object} string
// @Failure      500  {object} string
// @Router       /notifier/api/addressbook [post]
func (a *AddressBookController) HandleCreate(w http.ResponseWriter, r *http.Request) {
	a.HandleUpsertRaw(w, r, tools.UUIDGen())
}
