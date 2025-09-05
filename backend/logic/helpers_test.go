package logic

import (
	"notifier/repo"
	"notifier/tools"
	"testing"
	"time"
)

func newTestReminder(ty repo.ReminderType, wa []repo.WarningType, recipients []*tools.UUID) *repo.Reminder {
	r := &repo.Reminder{
		Id:          tools.UUIDGen(),
		Kind:        ty,
		Param:       0,
		WarningAt:   wa,
		Spec:        time.Now().Add(time.Hour * 24 * 8),
		Description: "Test",
		Recipients:  recipients,
	}

	return r
}

func newOneShotReminder(wa []repo.WarningType, recipients []*tools.UUID) *repo.Reminder {
	return newTestReminder(repo.OneShot, wa, recipients)
}

func TestRescheduleOneShot(t *testing.T) {
	martin := tools.UUIDGen()
	push := tools.UUIDGen()
	rem := newOneShotReminder([]repo.WarningType{repo.SameDay}, []*tools.UUID{martin, push})
	sch := NewGenericNotificationGenerator(false, oneShotRefTimeGen, tools.GenerateNotificationText)
	notifications, err := sch.Reschedule(rem)
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(notifications) != 2 {
		t.Errorf("Wrong number of notifications: %d", len(notifications))
	}

	rem = newOneShotReminder([]repo.WarningType{repo.SameDay, repo.EveningBefore, repo.WeekBefore}, []*tools.UUID{martin, push})
	notifications, err = sch.Reschedule(rem)
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(notifications) != 6 {
		t.Errorf("Wrong number of notifications: %d", len(notifications))
	}
}
