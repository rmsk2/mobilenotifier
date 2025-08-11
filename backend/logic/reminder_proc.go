package logic

import (
	"fmt"
	"log"
	"notifier/repo"
	"notifier/tools"
)

func ReminderTypeToGenerator(k repo.ReminderType) (NotificationGenerator, error) {
	switch k {
	case repo.OneShot:
		return NewGenericNotificationGenerator(false, oneShotRefTimeGen), nil
	case repo.Anniversary:
		return NewGenericNotificationGenerator(true, anniversaryRefTimeGen), nil
	case repo.Monthly:
		return NewGenericNotificationGenerator(true, monthlyRefTimeGen), nil
	default:
		return nil, fmt.Errorf("unknown reminder type: %d", k)
	}
}

func ProcessNewUuid(repoNotify repo.NotificationRepoWrite, repoReminder repo.ReminderRepoWrite, reminder *repo.Reminder) error {
	return ProcessOneUuid(repoNotify, repoReminder, reminder, true)
}

func ProcessOneUuid(repoNotify repo.NotificationRepoWrite, repoReminder repo.ReminderRepoWrite, reminder *repo.Reminder, forceReschedule bool) error {
	c, err := repoNotify.CountSiblings(reminder.Id)
	if err != nil {
		return err
	}

	if c != 0 {
		return fmt.Errorf("there are existing notifications for reminder '%s'", reminder.Id)
	}

	proc, err := ReminderTypeToGenerator(reminder.Kind)
	if err != nil {
		return err
	}

	if !(forceReschedule || proc.IsRescheduleNeeded(reminder)) {
		return repoReminder.Delete(reminder.Id)
	}

	newNotifications, err := proc.Reschedule(reminder)
	if err != nil {
		return err
	}

	for _, j := range newNotifications {
		err := repoNotify.Upsert(j)
		if err != nil {
			return err
		}
	}

	return nil
}

func ProcessExpired(dbl repo.DBSerializer, extLog *log.Logger, uuidsToProcess []string) {
	nRepo, rRepo := dbl.Lock()
	defer func() { dbl.Unlock() }()

	for _, j := range uuidsToProcess {
		uid, ok := tools.NewUuidFromString(j)
		if !ok {
			extLog.Printf("Unable to parse reminder uuid '%s'", j)
			continue
		}

		reminder, err := rRepo.Get(uid)
		if err != nil {
			extLog.Printf("Unable to read reminder uuid '%s': %v", uid, err)
			continue
		}

		if reminder == nil {
			extLog.Printf("Reminder uuid '%s' not found in repo", uid)
			continue
		}

		err = ProcessOneUuid(nRepo, rRepo, reminder, false)
		if err != nil {
			extLog.Printf("Unable to reschedule reminder '%s': %v", j, err)
		}
	}
}

func ChangeReminder(nWriteRepo repo.NotificationRepoWrite, writeRepo repo.ReminderRepoWrite, reminder *repo.Reminder) error {
	err := repo.ClearNotifications(nWriteRepo, reminder.Id)
	if err != nil {
		return fmt.Errorf("error clearing possibly existing notifications: %v", err)
	}

	err = writeRepo.Upsert(reminder)
	if err != nil {
		return fmt.Errorf("error creating/updating reminder: %v", err)
	}

	// ToDo: Attempt to cleanup DB if this fails
	err = ProcessNewUuid(nWriteRepo, writeRepo, reminder)
	if err != nil {
		return fmt.Errorf("error creating notifications for new/updated reminder: %v", err)
	}

	return nil
}

func RemoveReminder(nWriteRepo repo.NotificationRepoWrite, writeRepo repo.ReminderRepoWrite, id *tools.UUID) error {
	err := repo.ClearNotifications(nWriteRepo, id)
	if err != nil {
		return fmt.Errorf("error clearing possibly existing notifications: %v", err)
	}

	err = writeRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("error deleting reminder: %v", err)
	}

	return nil
}

func HasRecipient(recipientId *tools.UUID) repo.ReminderPredicate {
	return func(r *repo.Reminder) bool {
		for _, j := range r.Recipients {
			// ToDO: Enforce upper case
			if j == recipientId.String() {
				return true
			}
		}

		return false
	}
}

func TestAndRemoveRecipient(recipient string, recipients []string) ([]string, bool) {
	res := []string{}
	found := false

	for _, j := range recipients {
		if j != recipient {
			res = append(res, j)
		} else {
			found = true
		}
	}

	return res, found
}

func DeleteAddrBookEntry(nWriteRepo repo.NotificationRepoWrite, writeRepo repo.ReminderRepoWrite, addrBookWriteRepo repo.AddrBookWrite, addrEntryId *tools.UUID) error {
	affected, err := writeRepo.Filter(HasRecipient(addrEntryId))
	if err != nil {
		return err
	}

	for _, j := range affected {
		newRecipients, found := TestAndRemoveRecipient(addrEntryId.String(), j.Recipients)
		if found {
			if len(newRecipients) == 0 {
				err := RemoveReminder(nWriteRepo, writeRepo, j.Id)
				if err != nil {
					return err
				}
			} else {
				j.Recipients = newRecipients
				err := ChangeReminder(nWriteRepo, writeRepo, j)
				if err != nil {
					return err
				}
			}
		}
	}

	err = addrBookWriteRepo.Delete(addrEntryId)
	if err != nil {
		return err
	}

	return nil
}
