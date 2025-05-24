package warner

import (
	"log"
	"notifier/repo"
	"notifier/sms"
	"notifier/tools"
	"time"
)

type expiryInfo struct {
	uuid        *tools.UUID
	parent      *tools.UUID
	recipient   string
	description string
}

type warningGenerator struct {
	db       *repo.DBLocker
	sender   sms.SmsSender
	addrBook sms.SmsAddressBook
	ticker   *time.Ticker
	log      *log.Logger
}

func Start(l *repo.DBLocker, sender sms.SmsSender, addrBook sms.SmsAddressBook, t *time.Ticker, lg *log.Logger) {
	warner := warningGenerator{
		db:       l,
		sender:   sender,
		addrBook: addrBook,
		ticker:   t,
		log:      lg,
	}

	go func() {
		for {
			t := <-warner.ticker.C
			warner.ProcessTick(t)
		}
	}()
}

func (w *warningGenerator) Collect(refTime time.Time) []expiryInfo {
	res := []expiryInfo{}
	raw := w.db.Lock(false)
	defer func() { w.db.Unlock(false) }()

	repo := repo.NewBBoltNotificationRepo(raw)
	uuids, err := repo.GetExpired(refTime)
	if err != nil {
		w.log.Printf("Unable to determine expired notifications: %v", err)
		return []expiryInfo{}
	}

	for _, j := range uuids {
		info, err := repo.Get(j)
		if err != nil {
			log.Printf("Unable to retrieve info for notification id '%s'", j)
		} else {
			currentInfo := expiryInfo{
				uuid:        j,
				parent:      info.Parent,
				recipient:   info.Recipient,
				description: info.Description,
			}
			res = append(res, currentInfo)
		}
	}

	return res
}

func (w *warningGenerator) SendAndDeleteOne(info expiryInfo) {
	raw := w.db.Lock(true)
	defer func() { w.db.Unlock(true) }()

	repo := repo.NewBBoltNotificationRepo(raw)
	ok, err := w.addrBook.CheckRecipient(info.recipient)
	if err != nil {
		w.log.Printf("Unable to determine validity of recipient '%s': %v", info.recipient, err)
		return
	}

	if !ok {
		w.log.Printf("Recipeient '%s' in notification '%s' is invalid. Deleting notification", info.recipient, info.uuid)
		err = repo.Delete(info.uuid)
		if err != nil {
			w.log.Printf("Unable to delete notification '%s': %v", info.uuid, err)
		}
		return
	}

	err = w.sender.Send(info.recipient, info.description)
	if err != nil {
		w.log.Printf("Unable to send SMS to '%s' for notification '%s': %v", info.recipient, info.uuid, err)
		return
	}

	err = repo.Delete(info.uuid)
	if err != nil {
		w.log.Printf("Unable to delete notification '%s': %v", info.uuid, err)
	}
}

func (w *warningGenerator) SendAndDelete(expiredNotifications []expiryInfo) {
	for _, j := range expiredNotifications {
		w.SendAndDeleteOne(j)
	}
}

func (w *warningGenerator) ProcessTick(refTime time.Time) {
	w.log.Printf("Ticking at %v", refTime)

	res := w.Collect(refTime)
	w.SendAndDelete(res)
}
