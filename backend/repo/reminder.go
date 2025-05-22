package repo

import (
	"notifier/tools"
	"time"
)

type MetaKind int
type WarningType int

const (
	Anniversary MetaKind = iota + 1
	WeeklyEvent
	OneShot
)

const (
	MorningBefore WarningType = iota + 1
	NoonBefore
	EveningBefore
	WeekBefore
)

type Reminder struct {
	Id          *tools.UUID
	Parent      *tools.UUID
	WarningTime time.Time
	Description string
	Recipient   string
}

type MetaReminder struct {
	Id          *tools.UUID
	Kind        MetaKind
	Param       int
	WarningAt   []WarningType
	Spec        time.Time
	Description string
	Recipients  []string
}

//type ReminderGeneratorFunc func(time.Time, int) ([]*Reminder, error)

type ReminderPredicate func(r *Reminder) bool

type ReminderRepo interface {
	Upsert(r *Reminder) error
	Get(u *tools.UUID) (*Reminder, error)
	Delete(u *tools.UUID) error
	GetExpired() ([]*tools.UUID, error)
	CountSiblings(parent *tools.UUID) (int, error)
	Filter(p ReminderPredicate) ([]*tools.UUID, error)
}

type MetaReminderPredicate func(m *MetaReminder) bool

type MetaReminderRepo interface {
	Upsert(m *MetaReminder) error
	Get(u *tools.UUID) (*MetaReminder, error)
	Delete(u *tools.UUID) error
	Filter(p MetaReminderPredicate) ([]*tools.UUID, error)
}
