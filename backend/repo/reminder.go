package repo

import (
	"notifier/tools"
	"time"
)

type ReminderType int
type WarningType int

const (
	Anniversary ReminderType = iota + 1
	WeeklyEvent
	OneShot
)

const (
	MorningBefore WarningType = iota + 1
	NoonBefore
	EveningBefore
	WeekBefore
)

type Notification struct {
	Id          *tools.UUID
	Parent      *tools.UUID
	WarningTime time.Time
	Description string
	Recipient   string
}

type Reminder struct {
	Id          *tools.UUID
	Kind        ReminderType
	Param       int
	WarningAt   []WarningType
	Spec        time.Time
	Description string
	Recipients  []string
}

type NotificationPredicate func(r *Notification) bool

type NotificationRepoRead interface {
	Get(u *tools.UUID) (*Notification, error)
	GetExpired(time.Time) ([]*tools.UUID, error)
	CountSiblings(parent *tools.UUID) (int, error)
	Filter(p NotificationPredicate) ([]*tools.UUID, error)
}

type NotificationRepoWrite interface {
	NotificationRepoRead
	Upsert(n *Notification) error
	Delete(u *tools.UUID) error
}

type ReminderPredicate func(m *Reminder) bool

type ReminderRepoRead interface {
	Get(u *tools.UUID) (*Reminder, error)
	Filter(p ReminderPredicate) ([]*Reminder, error)
}

type ReminderRepoWrite interface {
	ReminderRepoRead
	Upsert(r *Reminder) error
	Delete(u *tools.UUID) error
}
