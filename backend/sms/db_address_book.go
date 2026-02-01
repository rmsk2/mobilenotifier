package sms

import (
	"fmt"
	"notifier/repo"
	"notifier/tools"
)

const TypeIFTTT = "IFTTT"
const TypeMail = "Mail"
const TypeDummy = "Dummy"
const TypeLocal = "local"

type RecipientInfo struct {
	Id          *tools.UUID `json:"id"`
	DisplayName string      `json:"display_name"`
}

type SmsAddressBook interface {
	ListRecipients() ([]RecipientInfo, error)
	GetSender(recipientId *tools.UUID) SmsSender
	AddSender(addrType string, s SmsSender)
	SetDefaultType(t string)
	CheckRecipient(r *tools.UUID) (bool, string, error)
	GetDefaultRecipientIds() []string
	GetAllAddressTypes() []string
}

func NewDBAddressBook(d repo.DBSerializer, g func(repo.DbType) *repo.BBoltAddrBookRepo) *DBAddressBook {
	gr := func(db repo.DbType) repo.AddrBookRead {
		return g(db)
	}

	return &DBAddressBook{
		db:          d,
		senders:     map[string]SmsSender{},
		defaultType: TypeIFTTT,
		genRead:     gr,
	}
}

type DBAddressBook struct {
	db          repo.DBSerializer
	senders     map[string]SmsSender
	defaultType string
	genRead     func(repo.DbType) repo.AddrBookRead
}

func (d *DBAddressBook) ListRecipients() ([]RecipientInfo, error) {
	readRepo := repo.LockAndGetRepoR(d.db, d.genRead)
	defer func() { d.db.RUnlock() }()

	recipients, err := readRepo.Filter(func(*repo.Recipient) bool { return true })
	if err != nil {
		return nil, fmt.Errorf("error getting all recipients: %v", err)
	}

	result := []RecipientInfo{}

	for _, j := range recipients {
		h := RecipientInfo{
			Id:          j.Id,
			DisplayName: j.DisplayName,
		}
		result = append(result, h)
	}

	return result, nil
}

func (d *DBAddressBook) GetSender(recipientId *tools.UUID) SmsSender {
	defaultSender := d.senders[d.defaultType]

	readRepo := repo.LockAndGetRepoR(d.db, d.genRead)
	defer func() { d.db.RUnlock() }()

	recipient, err := readRepo.Get(recipientId)
	if (err != nil) || (recipient == nil) {
		return defaultSender
	}

	sender, ok := d.senders[recipient.AddrType]
	if !ok {
		return defaultSender
	}

	return sender
}

func (d *DBAddressBook) CheckRecipient(r *tools.UUID) (bool, string, error) {
	readRepo := repo.LockAndGetRepoR(d.db, d.genRead)
	defer func() { d.db.RUnlock() }()

	recipient, err := readRepo.Get(r)
	if err != nil {
		return false, "", fmt.Errorf("unable to determine validity of recipient: '%s'", r)
	}

	if recipient == nil {
		return false, "", nil
	}

	return true, recipient.Address, nil
}

func (d *DBAddressBook) GetDefaultRecipientIds() []string {
	readRepo := repo.LockAndGetRepoR(d.db, d.genRead)
	defer func() { d.db.RUnlock() }()

	recipients, err := readRepo.Filter(func(r *repo.Recipient) bool {
		return r.IsDefault
	})
	if err != nil {
		return []string{}
	}

	result := []string{}

	for _, j := range recipients {
		result = append(result, j.Id.String())
	}

	return result
}

func (d *DBAddressBook) GetAllAddressTypes() []string {
	var allTypes []string = []string{}

	for i := range d.senders {
		allTypes = append(allTypes, i)
	}

	return allTypes
}

func (d *DBAddressBook) AddSender(addrType string, s SmsSender) {
	d.senders[addrType] = s
}

func (d *DBAddressBook) SetDefaultType(t string) {
	d.defaultType = t
}
