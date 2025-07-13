package repo

import (
	"notifier/tools"
	"time"
)

type ReminderType int
type WarningType int

const (
	Anniversary ReminderType = iota + 1
	OneShot
	Monthly
)

const (
	MorningBefore WarningType = iota + 1
	NoonBefore
	EveningBefore
	WeekBefore
	SameDay
)

type Notification struct {
	Id          *tools.UUID `json:"id"`
	Parent      *tools.UUID `json:"parent"`
	WarningTime time.Time   `json:"warning_time"`
	Description string      `json:"description"`
	Recipient   string      `json:"recipient"`
}

type Reminder struct {
	Id          *tools.UUID   `json:"id"`
	Kind        ReminderType  `json:"kind"`
	Param       int           `json:"param"`
	WarningAt   []WarningType `json:"warning_at"`
	Spec        time.Time     `json:"spec"`
	Description string        `json:"description"`
	Recipients  []string      `json:"recipients"`
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

func ClearNotifications(repoNotificationWrite NotificationRepoWrite, parentId *tools.UUID) error {
	uuids, err := repoNotificationWrite.Filter(func(n *Notification) bool {
		return n.Parent.IsEqual(parentId)
	})

	if err != nil {
		return err
	}

	for _, j := range uuids {
		err := repoNotificationWrite.Delete(j)
		if err != nil {
			return err
		}
	}

	return nil
}

func CountEntries(repo ReminderRepoRead) (int, error) {
	var elementCount int = 0
	_, err := repo.Filter(func(r *Reminder) bool {
		elementCount++
		return false
	})

	if err != nil {
		return -1, err
	}

	return elementCount, nil
}
