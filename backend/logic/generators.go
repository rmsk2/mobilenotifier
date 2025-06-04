package logic

import (
	"notifier/repo"
	"time"
)

type ReftimeGenerator func(*repo.Reminder, time.Time) time.Time

var RefTimeMap map[repo.ReminderType]ReftimeGenerator = map[repo.ReminderType]ReftimeGenerator{
	repo.Anniversary: anniversaryRefTimeGen,
	repo.OneShot:     oneShotRefTimeGen,
}

func oneShotRefTimeGen(r *repo.Reminder, now time.Time) time.Time {
	return r.Spec
}

func anniversaryRefTimeGen(r *repo.Reminder, now time.Time) time.Time {
	// Feb 29 is handled correctly by go, i.e. Feb 29 plus one year is Mar 01
	h := r.Spec.Local()
	refThisYear := time.Date(now.Year(), h.Month(), h.Day(), 0, 0, 0, 0, time.Local)
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
	refThisYear = time.Date(refThisYear.Year(), refThisYear.Month(), refThisYear.Day(), h.Hour(), h.Minute(), 0, 0, time.Local)

	return refThisYear
}

func weeklyRefTimeGen(r *repo.Reminder, now time.Time) time.Time {
	h := r.Spec.Local()
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

	now = time.Date(now.Year(), now.Month(), now.Day(), h.Hour(), h.Minute(), 0, 0, time.Local)
	refThisWeek = now.AddDate(0, 0, offset)

	return refThisWeek
}
