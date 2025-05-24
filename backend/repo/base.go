package repo

import (
	"fmt"
	"notifier/tools"
	"sync"

	bolt "go.etcd.io/bbolt"
)

type DBLocker struct {
	db    *bolt.DB
	mutex *sync.RWMutex
}

func (l *DBLocker) Lock(doWrite bool) *bolt.DB {
	if doWrite {
		l.mutex.Lock()
	} else {
		l.mutex.RLock()
	}

	return l.db
}

func (l *DBLocker) Unlock(doWrite bool) {
	if doWrite {
		l.mutex.Unlock()
	} else {
		l.mutex.RUnlock()
	}
}

func InitDB(openFlag *bool, boltPath string) (*DBLocker, *bolt.DB, error) {
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

	res := DBLocker{
		db:    db,
		mutex: new(sync.RWMutex),
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

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create buckets: %v", err)
	}

	return nil
}
