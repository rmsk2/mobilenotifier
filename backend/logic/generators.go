package logic

import (
	"notifier/repo"
	"time"
)

func oneShotRefTimeGen(r *repo.Reminder) time.Time {
	return r.Spec
}

func anniversaryRefTimeGen(r *repo.Reminder) time.Time {
	// ToDo: Handle Feb 29
	h := r.Spec.Local()
	now := time.Now()
	refThisYear := time.Date(now.Year(), h.Month(), h.Day(), 0, 0, 0, 0, time.Local)

	if refThisYear.Compare(now) == -1 {
		// In this year the anniversary is already in the past
		refThisYear = time.Date(now.Year()+1, h.Month(), h.Day(), 0, 0, 0, 0, time.Local)
	}

	return refThisYear
}

func weeklyRefTimeGen(r *repo.Reminder) time.Time {
	h := r.Spec.Local()
	now := time.Now()
	var refThisWeek time.Time
	var offset int

	if h.Weekday() < now.Weekday() {
		// In this week the day has already passed => nex week
		offset = 7 - (int(now.Weekday()) - int(h.Weekday()))
	} else {
		// In this week the day will occur in the future
		offset = int(h.Weekday()) - int(now.Weekday())
	}

	now = time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.Local)
	refThisWeek = now.AddDate(0, 0, -int(offset))

	return refThisWeek
}
