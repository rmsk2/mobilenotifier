package repo

import (
	"encoding/json"
	"fmt"
	"notifier/tools"

	bolt "go.etcd.io/bbolt"
)

func NewBBoltReminderRepo(db *bolt.DB) *BoltReminderRepo {
	return &BoltReminderRepo{
		db: db,
	}
}

type BoltReminderRepo struct {
	db *bolt.DB
}

func (b *BoltReminderRepo) Upsert(r *Reminder) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		// Store full Reminder
		b := tx.Bucket([]byte(bucketReminders))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketReminders)
		}

		data, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("unable to upsert reminder: %v", err)
		}

		err = b.Put(r.Id.AsSlice(), data)
		if err != nil {
			return fmt.Errorf("unable to upsert reminder: %v", err)
		}

		return err
	})

	return err
}

func (b *BoltReminderRepo) Get(u *tools.UUID) (*Reminder, error) {
	var res *Reminder = nil

	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketReminders))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketReminders)
		}

		v := b.Get(u.AsSlice())
		if v == nil {
			// value not found
			return nil
		}

		res = new(Reminder)

		err := json.Unmarshal(v, res)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to read reminder: %v", err)
	}

	return res, nil
}

func (b *BoltReminderRepo) Delete(u *tools.UUID) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		// Delete from reminders
		b := tx.Bucket([]byte(bucketReminders))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketReminders)
		}

		err := b.Delete(u.AsSlice())
		if err != nil {
			return fmt.Errorf("unable to delete notification: %v", err)
		}

		return err
	})

	return err
}

func (b *BoltReminderRepo) Filter(p ReminderPredicate) ([]*Reminder, error) {
	res := []*Reminder{}

	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketReminders))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketReminders)
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			entry := new(Reminder)
			err := json.Unmarshal(v, entry)
			if err != nil {
				return fmt.Errorf("unable to deserialize reminder: %v", err)
			}

			if p(entry) {
				res = append(res, entry)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to filter reminders: %v", err)
	}

	return res, nil
}
