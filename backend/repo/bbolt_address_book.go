package repo

import (
	"encoding/json"
	"fmt"
	"notifier/tools"

	bolt "go.etcd.io/bbolt"
)

type Recipient struct {
	DisplayName string     `json:"display_name"`
	Id          tools.UUID `json:"id"`
	Address     string     `json:"address"`
	AddrType    string     `json:"addr_type"`
	IsDefault   bool       `json:"is_default"`
}

type RecipientPredicate func(r *Recipient) bool

type AddrBookRead interface {
	Get(u *tools.UUID) (*Recipient, error)
	Filter(p RecipientPredicate) ([]*Recipient, error)
}

type AddrBookWrite interface {
	AddrBookRead
	Delete(u *tools.UUID) error
	Upsert(r *Recipient) error
}

func NewBBoltAddressBookRepo(d *bolt.DB) *BBoltAddrBookRepo {
	return &BBoltAddrBookRepo{
		db: d,
	}
}

type BBoltAddrBookRepo struct {
	db *bolt.DB
}

func (a *BBoltAddrBookRepo) Get(u *tools.UUID) (*Recipient, error) {
	var res *Recipient = nil

	err := a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketAddressBook))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketAddressBook)
		}

		v := b.Get(u.AsSlice())
		if v == nil {
			// value not found
			return nil
		}

		res = new(Recipient)

		err := json.Unmarshal(v, res)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to read addr book entry: %v", err)
	}

	return res, nil
}

//func (a *BBoltAddrBookRepo) SaveAddressBook(addrBook)

func (a *BBoltAddrBookRepo) Delete(u *tools.UUID) error {
	err := a.db.Update(func(tx *bolt.Tx) error {
		// Delete from address book
		b := tx.Bucket([]byte(bucketAddressBook))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketAddressBook)
		}

		err := b.Delete(u.AsSlice())
		if err != nil {
			return fmt.Errorf("unable to delete addr book entry: %v", err)
		}

		return err
	})

	return err
}

func (a *BBoltAddrBookRepo) Upsert(r *Recipient) error {
	err := a.db.Update(func(tx *bolt.Tx) error {
		// Store full address book entry
		b := tx.Bucket([]byte(bucketAddressBook))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketAddressBook)
		}

		data, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("unable to upsert address book entry: %v", err)
		}

		err = b.Put(r.Id.AsSlice(), data)
		if err != nil {
			return fmt.Errorf("unable to upsert address book entry: %v", err)
		}

		return err
	})

	return err
}

func (a *BBoltAddrBookRepo) Filter(p RecipientPredicate) ([]*Recipient, error) {
	res := []*Recipient{}

	err := a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketAddressBook))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketAddressBook)
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			entry := new(Recipient)
			err := json.Unmarshal(v, entry)
			if err != nil {
				return fmt.Errorf("unable to deserialize recipient: %v", err)
			}

			if p(entry) {
				res = append(res, entry)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to filter recipients: %v", err)
	}

	return res, nil
}
