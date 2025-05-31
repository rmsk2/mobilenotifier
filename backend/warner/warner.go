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
	db       repo.DBSerializer
	sender   sms.SmsSender
	addrBook sms.SmsAddressBook
	ticker   *time.Ticker
	log      *log.Logger
}

func Start(l repo.DBSerializer, sender sms.SmsSender, addrBook sms.SmsAddressBook, t *time.Ticker, lg *log.Logger) {
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
			warner.processTick(t)
		}
	}()
}

func (w *warningGenerator) collect(refTime time.Time) []expiryInfo {
	res := []expiryInfo{}
	readRepo, _ := w.db.RLock()
	defer func() { w.db.RUnlock() }()

	uuids, err := readRepo.GetExpired(refTime)
	if err != nil {
		w.log.Printf("Unable to determine expired notifications: %v", err)
		return []expiryInfo{}
	}

	for _, j := range uuids {
		info, err := readRepo.Get(j)
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

func (w *warningGenerator) sendAndDeleteOne(info expiryInfo) bool {
	writeRepo, _ := w.db.Lock()
	defer func() { w.db.Unlock() }()

	ok, err := w.addrBook.CheckRecipient(info.recipient)
	if err != nil {
		w.log.Printf("Unable to determine validity of recipient '%s': %v", info.recipient, err)
		return false
	}

	if !ok {
		w.log.Printf("Recipeient '%s' in notification '%s' is invalid. Deleting notification", info.recipient, info.uuid)
		err = writeRepo.Delete(info.uuid)
		if err != nil {
			w.log.Printf("Unable to delete notification '%s': %v", info.uuid, err)
			return false
		}
		return true
	}

	err = w.sender.Send(info.recipient, info.description)
	if err != nil {
		w.log.Printf("Unable to send SMS to '%s' for notification '%s': %v", info.recipient, info.uuid, err)
		return false
	}

	w.log.Printf("Message sent to '%s' for notification '%s'", info.recipient, info.uuid)

	err = writeRepo.Delete(info.uuid)
	if err != nil {
		w.log.Printf("Unable to delete notification '%s': %v", info.uuid, err)
		return false
	}

	w.log.Printf("Notification '%s' deleted", info.uuid)

	return true
}

func (w *warningGenerator) determineChildlessParents(affectedParents map[*tools.UUID]bool) []string {
	res := []string{}

	readRepo, _ := w.db.RLock()
	defer func() { w.db.RUnlock() }()

	for i := range affectedParents {
		count, err := readRepo.CountSiblings(i)
		if err != nil {
			log.Printf("Problem: Unable to determine child count for parent '%s'. This could create a dead reminder", i)
			continue
		}

		if count == 0 {
			res = append(res, i.String())
		}
	}

	return res
}

func (w *warningGenerator) processTick(refTime time.Time) {
	w.log.Printf("Ticking at %v", refTime)
	affectedParents := map[*tools.UUID]bool{}

	expiredNotifications := w.collect(refTime)

	for _, j := range expiredNotifications {
		if w.sendAndDeleteOne(j) {
			affectedParents[j.parent] = true
		}
	}

	// Prevent locking of database if there is nothing to do
	if len(affectedParents) == 0 {
		return
	}

	_ = w.determineChildlessParents(affectedParents)
}
