package logic

import (
	"fmt"
	"log"
	"notifier/repo"
	"notifier/sms"
	"notifier/tools"
	"time"
)

const metricsTicks = "ticks"

type AddMetricsEvent func(string)

type expiryInfo struct {
	uuid        *tools.UUID
	parent      *tools.UUID
	recipient   *tools.UUID
	description string
}

type warningGenerator struct {
	db             repo.DBSerializer
	addrBook       sms.SmsAddressBook
	ticker         *time.Ticker
	log            *log.Logger
	metricCallback AddMetricsEvent
}

func StartWarner(l repo.DBSerializer, addrBook sms.SmsAddressBook, t *time.Ticker, lg *log.Logger, m AddMetricsEvent) {
	warner := warningGenerator{
		db:             l,
		addrBook:       addrBook,
		ticker:         t,
		log:            lg,
		metricCallback: m,
	}

	go func() {
		for {
			t := <-warner.ticker.C
			warner.metricCallback(metricsTicks)
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
			w.log.Printf("Unable to retrieve info for notification id '%s'", j)
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
	// First obtain lock on Reminder and Notification store and only after
	// that get a lock on AddressBook. This prevents deadlocks.
	writeRepo, _ := w.db.Lock()
	defer func() { w.db.Unlock() }()

	ok, address, err := w.addrBook.CheckRecipient(info.recipient)
	if err != nil {
		w.log.Printf("Unable to determine validity of recipient '%s': %v", info.recipient, err)
		return false
	}

	if !ok {
		w.log.Printf("Recipient '%s' in notification '%s' is invalid. Deleting notification", info.recipient, info.uuid)
		err = writeRepo.Delete(info.uuid)
		if err != nil {
			w.log.Printf("Unable to delete notification '%s': %v", info.uuid, err)
			return false
		}
		return true
	}

	sender := w.addrBook.GetSender(info.recipient)

	err = sender.Send(address, info.description)
	if err != nil {
		w.log.Printf("Unable to send SMS to '%s' for notification '%s': %v", info.recipient, info.uuid, err)
		return false
	}

	w.log.Printf("Message sent to '%s' for notification '%s'", info.recipient, info.uuid)

	if w.metricCallback != nil {
		w.metricCallback(tools.NotificationSent)
		w.metricCallback(fmt.Sprintf("%s:%s", address, info.recipient.String()))
		w.metricCallback(sender.GetName())
	}

	err = writeRepo.Delete(info.uuid)
	if err != nil {
		w.log.Printf("Unable to delete notification '%s': %v", info.uuid, err)
		return false
	}

	w.log.Printf("Notification '%s' deleted", info.uuid)

	return true
}

func (w *warningGenerator) determineChildlessParents(affectedParents map[string]bool) []string {
	res := []string{}

	readRepo, _ := w.db.RLock()
	defer func() { w.db.RUnlock() }()

	for i := range affectedParents {
		u, _ := tools.NewUuidFromString(i)
		count, err := readRepo.CountSiblings(u)
		if err != nil {
			w.log.Printf("Problem: Unable to determine child count for parent '%s'. This could create a dead reminder", i)
			continue
		}

		if count == 0 {
			res = append(res, i)
		}
	}

	return res
}

func (w *warningGenerator) processTick(refTime time.Time) {
	w.log.Printf("Ticking at %v", refTime)
	affectedParents := map[string]bool{}

	expiredNotifications := w.collect(refTime)

	for _, j := range expiredNotifications {
		if w.sendAndDeleteOne(j) {
			affectedParents[j.parent.String()] = true
		}
	}

	// Prevent locking of database if there is nothing to do
	if len(affectedParents) == 0 {
		return
	}

	remindersToReschedule := w.determineChildlessParents(affectedParents)
	ProcessExpired(w.db, w.log, remindersToReschedule)
}
