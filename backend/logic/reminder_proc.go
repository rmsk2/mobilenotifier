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
	case repo.WeeklyEvent:
		return NewGenericNotificationGenerator(true, weeklyRefTimeGen), nil
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
