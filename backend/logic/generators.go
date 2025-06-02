package logic

import (
	"notifier/repo"
	"time"
)

func oneShotRefTimeGen(r *repo.Reminder) time.Time {
	return r.Spec
}

func anniversaryRefTimeGen(r *repo.Reminder) time.Time {
	// Feb 29 is handled correctly by go, i.e. Feb 29 plus one year is Mar 01
	h := r.Spec.Local()
	now := time.Now()
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

	return refThisYear
}

func weeklyRefTimeGen(r *repo.Reminder) time.Time {
	h := r.Spec.Local()
	now := time.Now()
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

	now = time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.Local)
	refThisWeek = now.AddDate(0, 0, offset)

	return refThisWeek
}
