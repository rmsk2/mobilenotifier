package logic

import (
	"notifier/repo"
	"notifier/tools"
	"time"
)

type NotifcationMsgGenerator func(string, int, int, string) string

type NotificationGenerator interface {
	IsRescheduleNeeded(*repo.Reminder) bool
	Reschedule(*repo.Reminder) ([]*repo.Notification, error)
}

type OffsetGenerator func(time.Time, int) (time.Time, string)

type GenericNotificationGenerator struct {
	rescheduleNeeded bool
	offsetGens       map[repo.WarningType]OffsetGenerator
	genRefTime       ReftimeGenerator
	genNotifText     NotifcationMsgGenerator
}

func toYesterday(t time.Time) time.Time {
	help := t.In(tools.ClientTZ())
	return help.AddDate(0, 0, -1).UTC()
}

func morningBefore(t time.Time, p int) (time.Time, string) {
	t = toYesterday(t).In(tools.ClientTZ())
	return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, tools.ClientTZ()).UTC(), tools.MsgTextTomorrow
}

func noonBefore(t time.Time, p int) (time.Time, string) {
	t = toYesterday(t).In(tools.ClientTZ())
	return time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, tools.ClientTZ()).UTC(), tools.MsgTextTomorrow
}

func eveningBefore(t time.Time, p int) (time.Time, string) {
	t = toYesterday(t).In(tools.ClientTZ())
	return time.Date(t.Year(), t.Month(), t.Day(), 18, 0, 0, 0, tools.ClientTZ()).UTC(), tools.MsgTextTomorrow
}

func weekBefore(t time.Time, p int) (time.Time, string) {
	t = t.In(tools.ClientTZ()).AddDate(0, 0, -7)
	return time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, tools.ClientTZ()).UTC(), tools.MsgTextInSevenDays
}

func sameDay(t time.Time, p int) (time.Time, string) {
	duration := time.Hour * time.Duration(p&31)
	return t.Add(-duration), tools.MsgTextToday
}

func NewGenericNotificationGenerator(r bool, g ReftimeGenerator, txtGen NotifcationMsgGenerator) *GenericNotificationGenerator {
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
		genNotifText:     txtGen,
	}

	return res
}

func (g *GenericNotificationGenerator) IsRescheduleNeeded(r *repo.Reminder) bool {
	return g.rescheduleNeeded
}

type offsetTuple struct {
	t      time.Time
	prefix string
}

func (g *GenericNotificationGenerator) Reschedule(r *repo.Reminder) ([]*repo.Notification, error) {
	res := []*repo.Notification{}
	times := []offsetTuple{}
	refNowUtc := time.Now().UTC()

	refTime := g.genRefTime(r, refNowUtc)

	for _, t := range r.WarningAt {
		ti, msgPrefix := g.offsetGens[t](refTime, r.Param)
		// Do not create notifications for a point in time which lies in the past
		if ti.After(refNowUtc) {
			h := offsetTuple{
				t:      ti,
				prefix: msgPrefix,
			}
			times = append(times, h)
		}
	}

	for _, i := range r.Recipients {
		for _, j := range times {
			eventLocalTime := r.Spec.In(tools.ClientTZ())
			n := new(repo.Notification)
			n.Id = tools.UUIDGen()
			n.Parent = r.Id
			n.Description = g.genNotifText(j.prefix, eventLocalTime.Hour(), eventLocalTime.Minute(), r.Description)
			n.WarningTime = j.t
			n.Recipient = i

			res = append(res, n)
		}
	}

	return res, nil
}
