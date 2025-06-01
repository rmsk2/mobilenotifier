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

func CloneForEachRecipient(recipients []string, n *repo.Notification) []*repo.Notification {
	res := []*repo.Notification{}

	for _, j := range recipients {
		h := new(repo.Notification)

		h.Id = tools.UUIDGen()
		h.Parent = n.Parent
		h.WarningTime = n.WarningTime
		h.Description = n.Description
		h.Recipient = j

		res = append(res, h)
	}

	return res
}

func CloneForEachWarningTime(types []repo.WarningType, warningTime time.Time, n *repo.Notification) []*repo.Notification {
	return []*repo.Notification{n}
}
