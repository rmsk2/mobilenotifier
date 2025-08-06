package repo

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"notifier/tools"
	"time"

	bolt "go.etcd.io/bbolt"
)

func NewBBoltNotificationRepo(db *bolt.DB) *BoltNotificationRepo {
	return &BoltNotificationRepo{
		db: db,
	}
}

type BoltNotificationRepo struct {
	db *bolt.DB
}

func int64ToBigEndian(i int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, (uint64)(i))

	return buf
}

func (b *BoltNotificationRepo) Upsert(n *Notification) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		// Store full notification
		b := tx.Bucket([]byte(bucketNotifications))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketNotifications)
		}

		data, err := json.Marshal(n)
		if err != nil {
			return fmt.Errorf("unable to upsert notification: %v", err)
		}

		err = b.Put(n.Id.AsSlice(), data)
		if err != nil {
			return fmt.Errorf("unable to upsert notification: %v", err)
		}

		// Store expiry time
		b = tx.Bucket([]byte(bucketExpiryTimes))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketExpiryTimes)
		}

		unixTimeStamp := n.WarningTime.Unix()

		err = b.Put(n.Id.AsSlice(), int64ToBigEndian(unixTimeStamp))
		if err != nil {
			return fmt.Errorf("unable to upsert notification: %v", err)
		}

		// Store parent
		b = tx.Bucket([]byte(bucketParents))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketParents)
		}

		err = b.Put(n.Id.AsSlice(), n.Parent.AsSlice())
		if err != nil {
			return fmt.Errorf("unable to upsert notification: %v", err)
		}

		return err
	})

	return err
}

func (b *BoltNotificationRepo) Get(u *tools.UUID) (*Notification, error) {
	var res *Notification = nil

	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketNotifications))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketNotifications)
		}

		v := b.Get(u.AsSlice())
		if v == nil {
			// value not found
			return nil
		}

		res = new(Notification)

		err := json.Unmarshal(v, res)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to read notification: %v", err)
	}

	return res, nil
}

func (b *BoltNotificationRepo) Delete(u *tools.UUID) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		// Delete from full notification
		b := tx.Bucket([]byte(bucketNotifications))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketNotifications)
		}

		err := b.Delete(u.AsSlice())
		if err != nil {
			return fmt.Errorf("unable to delete notification: %v", err)
		}

		// Delete from expiry time
		b = tx.Bucket([]byte(bucketExpiryTimes))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketExpiryTimes)
		}

		err = b.Delete(u.AsSlice())
		if err != nil {
			return fmt.Errorf("unable to delete notification: %v", err)
		}

		// Delete from parent
		b = tx.Bucket([]byte(bucketParents))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketParents)
		}

		err = b.Delete(u.AsSlice())
		if err != nil {
			return fmt.Errorf("unable to delete notification: %v", err)
		}

		return err
	})

	return err
}

func (b *BoltNotificationRepo) GetExpired(refTime time.Time) ([]*tools.UUID, error) {
	res := []*tools.UUID{}

	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketExpiryTimes))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketExpiryTimes)
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			if len(v) != 8 {
				return fmt.Errorf("illegal data length: %d", len(v))
			}

			if refTime.Unix() >= (int64)(binary.BigEndian.Uint64(v)) {
				u, ok := tools.NewUuidFromSlice(k)
				if !ok {
					return fmt.Errorf("illegal uid format: %x", v)
				}

				res = append(res, u)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to determine expired notifications: %v", err)
	}

	return res, nil
}

func (b *BoltNotificationRepo) CountSiblings(parent *tools.UUID) (int, error) {
	res := 0

	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketParents))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketParents)
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			u, ok := tools.NewUuidFromSlice(v)
			if !ok {
				return fmt.Errorf("illegal uuid format: %x", v)
			}

			if u.IsEqual(parent) {
				res++
			}
		}

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("unable to determine siblings: %v", err)
	}

	return res, nil
}

func (b *BoltNotificationRepo) Filter(p NotificationPredicate) ([]*tools.UUID, error) {
	res := []*tools.UUID{}

	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketNotifications))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketNotifications)
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			entry := new(Notification)
			err := json.Unmarshal(v, entry)
			if err != nil {
				return fmt.Errorf("unable to deserialize notification: %v", err)
			}

			if p(entry) {
				res = append(res, entry.Id)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to filter notifications: %v", err)
	}

	return res, nil
}
