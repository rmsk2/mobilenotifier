package logic

import (
	"notifier/repo"
	"notifier/tools"
	"time"
)

/*
type Notification struct {
	Id          *tools.UUID
	Parent      *tools.UUID
	WarningTime time.Time
	Description string
	Recipient   string
}
*/

type NotificationGenerator interface {
	IsRescheduleNeeded(*repo.Reminder) bool
	Reschedule(*repo.Reminder) ([]*repo.Notification, error)
}

type ReftimeGenerator func(*repo.Reminder) time.Time
type OffsetGenerator func(time.Time) time.Time

type GenericNotificationGenerator struct {
	rescheduleNeeded bool
	offsetGens       map[repo.WarningType]OffsetGenerator
	genRefTime       ReftimeGenerator
}

func toYesterday(t time.Time) time.Time {
	return t.Local().AddDate(0, 0, -1)
}

func morningBefore(t time.Time) time.Time {
	t = toYesterday(t)
	return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, time.Local)
}

func noonBefore(t time.Time) time.Time {
	t = toYesterday(t)
	return time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.Local)
}

func eveningBefore(t time.Time) time.Time {
	t = toYesterday(t)
	return time.Date(t.Year(), t.Month(), t.Day(), 18, 0, 0, 0, time.Local)
}

func weekBefore(t time.Time) time.Time {
	t = t.Local().AddDate(0, 0, -7)
	return time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.Local)
}

func sameDay(t time.Time) time.Time {
	return t.Local()
}

func NewGenericNotificationGenerator(r bool, g ReftimeGenerator) *GenericNotificationGenerator {
	offGens := map[repo.WarningType]OffsetGenerator{}

	offGens[repo.MorningBefore] = morningBefore
	offGens[repo.NoonBefore] = noonBefore
	offGens[repo.EveningBefore] = eveningBefore
	offGens[repo.WeekBefore] = weekBefore
	offGens[repo.SameDay] = sameDay

	res := &GenericNotificationGenerator{
		rescheduleNeeded: r,
		genRefTime:       g,
		offsetGens:       offGens,
	}

	return res
}

func (g *GenericNotificationGenerator) IsRescheduleNeeded(r *repo.Reminder) bool {
	return g.rescheduleNeeded
}

func (g *GenericNotificationGenerator) Reschedule(r *repo.Reminder) ([]*repo.Notification, error) {
	res := []*repo.Notification{}
	times := []time.Time{}

	refTime := g.genRefTime(r)

	for _, t := range r.WarningAt {
		times = append(times, g.offsetGens[t](refTime))
	}

	for _, i := range r.Recipients {
		for _, j := range times {
			n := new(repo.Notification)
			n.Id = tools.UUIDGen()
			n.Parent = r.Id
			n.Description = r.Description
			n.WarningTime = j
			n.Recipient = i

			res = append(res, n)
		}
	}

	return res, nil
}
