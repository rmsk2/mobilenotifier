package repo

import (
	"fmt"
	"notifier/tools"
	"sync"

	bolt "go.etcd.io/bbolt"
)

const bucketNotifications = "NOTIFICATIONS"
const bucketExpiryTimes = "EXPIRIES"
const bucketParents = "PARENTS"
const bucketReminders = "REMINDERS"
const bucketAddressBook = "ADDRESSBOOK"

type DbType *bolt.DB

type DBSerializer interface {
	RLock() (NotificationRepoRead, ReminderRepoRead)
	RUnlock()
	Lock() (NotificationRepoWrite, ReminderRepoWrite)
	Unlock()
	GetRaw() interface{}
}

type BoltDBLocker struct {
	db    DbType
	mutex *sync.RWMutex
}

func LockAndGetRepoRW[T, U any](l DBSerializer, generator func(U) T) T {
	l.Lock()

	return generator(l.GetRaw().(U))
}

func LockAndGetRepoR[T, U any](l DBSerializer, generator func(U) T) T {
	l.RLock()

	return generator(l.GetRaw().(U))
}

func (l *BoltDBLocker) Lock() (NotificationRepoWrite, ReminderRepoWrite) {
	l.mutex.Lock()

	return NewBBoltNotificationRepo(l.db), NewBBoltReminderRepo(l.db)
}

func (l *BoltDBLocker) Unlock() {
	l.mutex.Unlock()
}

func (l *BoltDBLocker) GetRaw() interface{} {
	return l.db
}

func (l *BoltDBLocker) RLock() (NotificationRepoRead, ReminderRepoRead) {
	l.mutex.RLock()

	return NewBBoltNotificationRepo(l.db), NewBBoltReminderRepo(l.db)
}

func (l *BoltDBLocker) RUnlock() {
	l.mutex.RUnlock()
}

func InitDB(openFlag *bool, boltPath string, dbMutex *sync.RWMutex) (*BoltDBLocker, *bolt.DB, error) {
	db, err := bolt.Open(boltPath, 0600, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to open database file %s: %v", boltPath, err)
	}

	*openFlag = true
	tools.InstallSignalHandler(db, openFlag)

	err = CreateBuckets(db)
	if err != nil {
		db.Close()
		*openFlag = false
		return nil, nil, fmt.Errorf("unable to create buckets in database file %s: %v", boltPath, err)
	}

	res := BoltDBLocker{
		db:    db,
		mutex: dbMutex,
	}

	return &res, db, nil
}

func CreateBuckets(db *bolt.DB) error {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketNotifications))
		if err != nil {
			return fmt.Errorf("error creating bucket for notifications: %v", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte(bucketExpiryTimes))
		if err != nil {
			return fmt.Errorf("error creating bucket for expirytimes: %v", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte(bucketParents))
		if err != nil {
			return fmt.Errorf("error creating bucket for parent ids: %v", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte(bucketReminders))
		if err != nil {
			return fmt.Errorf("error creating bucket for reminders: %v", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte(bucketAddressBook))
		if err != nil {
			return fmt.Errorf("error creating bucket for the address book: %v", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create buckets: %v", err)
	}

	return nil
}
