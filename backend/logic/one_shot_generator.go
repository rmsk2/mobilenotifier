package logic

import (
	"notifier/repo"
	"notifier/tools"
)

func NewOneShotGenerator() NotificationGenerator {
	return &OneShotGenerator{}
}

type OneShotGenerator struct {
}

func (o *OneShotGenerator) IsReschudleNeeded(r *repo.Reminder) bool {
	return false
}

func (o *OneShotGenerator) Reschedule(r *repo.Reminder) ([]*repo.Notification, error) {
	template := &repo.Notification{
		Id:          tools.UUIDGen(),
		Parent:      r.Id,
		WarningTime: r.Spec,
		Description: r.Description,
		Recipient:   r.Recipients[0],
	}

	res := CloneForEachRecipient(r.Recipients, template)

	return res, nil
}
