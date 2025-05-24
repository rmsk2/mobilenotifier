package repo

import (
	"fmt"
	"notifier/tools"
	"os"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

const pathTestDb = "testbold.bin"
const Uuid1 = "D20F95D6-3339-40FF-8E4E-B2F6AC439D06"
const Uuid2 = "3ACC4A36-5F13-49DC-A795-AA8553FEBAAE"
const Uuid3 = "F27834D6-3339-40FF-8E4E-B2F6AC439D06"

func Test1(t *testing.T) {
	os.Remove(pathTestDb)

	db, err := bolt.Open(pathTestDb, 0600, nil)
	if err != nil {
		t.Fatalf("Unable to open database file %s: %v\n", pathTestDb, err)
	}
	defer func() {
		db.Close()
		fmt.Println("bbolt DB closed")
		os.Remove(pathTestDb)
	}()

	err = CreateBuckets(db)
	if err != nil {
		t.Errorf("Creating buckets failed: %v", err)
		return
	}

	var r NotificationRepo = NewBBoltNotificationRepo(db)

	testUUID1, _ := tools.NewUuidFromString(Uuid1)
	testUUID2, _ := tools.NewUuidFromString(Uuid2)
	testUUID3, _ := tools.NewUuidFromString(Uuid3)

	n := Notification{
		Id:          testUUID1,
		Parent:      testUUID2,
		WarningTime: time.Now().UTC(),
		Description: "Test notification",
		Recipient:   "martin",
	}

	err = r.Upsert(&n)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
		return
	}

	n2 := Notification{
		Id:          testUUID3,
		Parent:      testUUID2,
		WarningTime: time.Now().UTC(),
		Description: "Test notification",
		Recipient:   "martin2",
	}

	err = r.Upsert(&n2)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
		return
	}

	d, err := r.Get(n2.Id)
	if err != nil {
		t.Errorf("Getting id %sfailed: %v", n2.Id, err)
		return
	}

	if d.Recipient != "martin2" {
		t.Errorf("Wrong values retrieved: %s", d.Recipient)
		return
	}

	c, err := r.CountSiblings(n.Parent)
	if err != nil {
		t.Errorf("Counting siblings failed: %v", err)
		return
	}

	if c != 2 {
		t.Errorf("Wrong number of siblings: %d", c)
	}

	err = r.Delete(n.Id)
	if err != nil {
		t.Errorf("Deleting uuid %s failed: %v", n.Id, err)
		return
	}

	c, err = r.CountSiblings(n.Parent)
	if err != nil {
		t.Errorf("Counting siblings failed: %v", err)
		return
	}

	if c != 1 {
		t.Errorf("Wrong number of siblings: %d", c)
	}

	res, err := r.GetExpired(time.Now())
	if err != nil {
		t.Errorf("Get expired notfications failed: %v", err)
		return
	}

	if len(res) != 1 {
		t.Errorf("Wrong number of expired notifications: %d", len(res))
	}

	if !res[0].IsEqual(n2.Id) {
		t.Errorf("Id of expired notification is wrong: %s", res[0])
	}

	err = r.Delete(n2.Id)
	if err != nil {
		t.Errorf("Deleting uuid %s failed: %v", n2.Id, err)
		return
	}

	res, err = r.GetExpired(time.Now())
	if err != nil {
		t.Errorf("Get expired notfications failed: %v", err)
		return
	}

	if len(res) != 0 {
		t.Errorf("Wrong number of expired notifications: %d", len(res))
	}

	c, err = r.CountSiblings(n.Parent)
	if err != nil {
		t.Errorf("Counting siblings failed: %v", err)
		return
	}

	if c != 0 {
		t.Errorf("Wrong number of siblings: %d", c)
	}
}

func Test2(t *testing.T) {
	os.Remove(pathTestDb)

	db, err := bolt.Open(pathTestDb, 0600, nil)
	if err != nil {
		t.Fatalf("Unable to open database file %s: %v\n", pathTestDb, err)
	}
	defer func() {
		db.Close()
		fmt.Println("bbolt DB closed")
		os.Remove(pathTestDb)
	}()

	err = CreateBuckets(db)
	if err != nil {
		t.Errorf("Creating buckets failed: %v", err)
		return
	}

	var r NotificationRepo = NewBBoltNotificationRepo(db)

	testUUID1, _ := tools.NewUuidFromString(Uuid1)
	testUUID2, _ := tools.NewUuidFromString(Uuid2)
	testUUID3, _ := tools.NewUuidFromString(Uuid3)

	n := Notification{
		Id:          testUUID1,
		Parent:      testUUID2,
		WarningTime: time.Now().UTC(),
		Description: "Test notification",
		Recipient:   "martin",
	}

	err = r.Upsert(&n)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
		return
	}

	n2 := Notification{
		Id:          testUUID3,
		Parent:      testUUID2,
		WarningTime: time.Now().UTC(),
		Description: "Test notification",
		Recipient:   "martin2",
	}

	err = r.Upsert(&n2)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
		return
	}

	res, err := r.Filter(func(n *Notification) bool {
		return n.Recipient == "martin2"
	})
	if err != nil {
		t.Errorf("Filtering failed: %v", err)
		return
	}

	if len(res) != 1 {
		t.Errorf("Filtering failed. Wrong number of results: %d", len(res))
		return
	}

	if !res[0].IsEqual(n2.Id) {
		t.Errorf("Filtering failed. Wrong number results: %s", res[0])
		return
	}
}

func Test3(t *testing.T) {
	os.Remove(pathTestDb)

	db, err := bolt.Open(pathTestDb, 0600, nil)
	if err != nil {
		t.Fatalf("Unable to open database file %s: %v\n", pathTestDb, err)
	}
	defer func() {
		db.Close()
		fmt.Println("bbolt DB closed")
		os.Remove(pathTestDb)
	}()

	err = CreateBuckets(db)
	if err != nil {
		t.Errorf("Creating buckets failed: %v", err)
		return
	}

	var r ReminderRepo = NewBBoltReminderRepo(db)

	testUUID1, _ := tools.NewUuidFromString(Uuid1)
	testUUID2, _ := tools.NewUuidFromString(Uuid2)

	n := Reminder{
		Id:          testUUID1,
		Kind:        OneShot,
		Param:       0,
		Spec:        time.Now().UTC(),
		WarningAt:   []WarningType{MorningBefore, NoonBefore},
		Description: "Test Reminder",
		Recipients:  []string{"martin"},
	}

	err = r.Upsert(&n)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
		return
	}

	n2 := Reminder{
		Id:          testUUID2,
		Kind:        OneShot,
		Param:       0,
		Spec:        time.Now().UTC(),
		WarningAt:   []WarningType{MorningBefore, NoonBefore},
		Description: "Test Reminder",
		Recipients:  []string{"martin2"},
	}

	err = r.Upsert(&n2)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
		return
	}

	n3, err := r.Get(n2.Id)
	if err != nil {
		t.Errorf("Getting element failed: %v", err)
		return
	}

	if (len(n3.Recipients) != 1) || (n3.Recipients[0] != "martin2") {
		t.Errorf("Data wrong")
	}

	count := 0

	res, err := r.Filter(func(n *Reminder) bool {
		count++
		return true
	})
	if err != nil {
		t.Errorf("Filtering failed: %v", err)
		return
	}

	if len(res) != 2 {
		t.Errorf("Filtering failed. Wrong number of results: %d", len(res))
		return
	}

	if count != 2 {
		t.Errorf("Filtering failed. Wrong number results: %d", count)
		return
	}

	err = r.Delete(n2.Id)
	if err != nil {
		t.Errorf("Deleting failed: %v", err)
		return
	}

	count = 0
	res, err = r.Filter(func(n *Reminder) bool {
		count++
		return true
	})
	if err != nil {
		t.Errorf("Filtering failed: %v", err)
		return
	}

	if len(res) != 1 {
		t.Errorf("Filtering failed. Wrong number of results: %d", len(res))
		return
	}

	if count != 1 {
		t.Errorf("Filtering failed. Wrong number results: %d", count)
		return
	}
}
