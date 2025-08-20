package logic

import (
	"notifier/repo"
	"notifier/tools"
	"time"
)

type ReftimeGenerator func(*repo.Reminder, time.Time) time.Time

var RefTimeMap map[repo.ReminderType]ReftimeGenerator = map[repo.ReminderType]ReftimeGenerator{
	repo.Anniversary: anniversaryRefTimeGen,
	repo.OneShot:     oneShotRefTimeGen,
	repo.Monthly:     monthlyRefTimeGen,
}

func oneShotRefTimeGen(r *repo.Reminder, now time.Time) time.Time {
	return r.Spec
}

// Calculates the next occurrance of the event defined by r.Spec in the clients timezone
func anniversaryRefTimeGen(r *repo.Reminder, n time.Time) time.Time {
	// Feb 29 is handled correctly by go, i.e. Feb 29 plus one year is Mar 01
	h := r.Spec.In(tools.ClientTZ())
	now := n.In(tools.ClientTZ())
	// A yearly event which is created on the day it occurs is scheduled in this year if the
	// event is still in the future relative to the current time given in parameter n.
	refThisYear := time.Date(now.Year(), h.Month(), h.Day(), h.Hour(), h.Minute(), 0, 0, tools.ClientTZ())
	var offset int

	switch {
	case refThisYear.Compare(now) == -1:
		offset = 1
	case refThisYear.Compare(now) == 0:
		offset = 1
	default:
		offset = 0
	}

	refThisYear = refThisYear.AddDate(offset, 0, 0)
	refThisYear = time.Date(refThisYear.Year(), refThisYear.Month(), refThisYear.Day(), h.Hour(), h.Minute(), 0, 0, tools.ClientTZ())

	return refThisYear.UTC()
}

// Calculates the next occurrance of the event defined by r.Spec in the clients timezone
func monthlyRefTimeGen(r *repo.Reminder, n time.Time) time.Time {
	h := r.Spec.In(tools.ClientTZ())
	now := n.In(tools.ClientTZ())
	// When creating a monthly event the desired hour and minute are taken into account in order to decide whether
	// to initially schedule the event in the current or the following month
	refThisMonth := time.Date(now.Year(), now.Month(), h.Day(), h.Hour(), h.Minute(), 0, 0, tools.ClientTZ())
	var offset int

	switch {
	case refThisMonth.Compare(now) == -1:
		offset = 1
	case refThisMonth.Compare(now) == 0:
		offset = 1
	default:
		offset = 0
	}

	refThisMonth = refThisMonth.AddDate(0, offset, 0)

	return refThisMonth.UTC()
}

// Calculates the next occurrance of the event defined by r.Spec in the clients timezone
func weeklyRefTimeGen(r *repo.Reminder, n time.Time) time.Time {
	h := r.Spec.In(tools.ClientTZ())
	now := n.In(tools.ClientTZ())
	var refThisWeek time.Time
	var offset int

	switch {
	case h.Weekday() == now.Weekday():
		offset = 7
	case h.Weekday() < now.Weekday():
		offset = 7 - (int(now.Weekday()) - int(h.Weekday()))
	default:
		offset = int(h.Weekday()) - int(now.Weekday())
	}

	now = time.Date(now.Year(), now.Month(), now.Day(), h.Hour(), h.Minute(), 0, 0, tools.ClientTZ())
	refThisWeek = now.AddDate(0, 0, offset)

	return refThisWeek.UTC()
}
