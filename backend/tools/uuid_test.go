package tools

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

type SerTest struct {
	Id1  *UUID
	Id2  *UUID
	Egal time.Time
}

func Test1(t *testing.T) {
	t1 := SerTest{
		Id1:  UUIDGen(),
		Id2:  nil,
		Egal: time.Now(),
	}

	res, err := json.MarshalIndent(&t1, "", "  ")
	if err != nil {
		t.Errorf("Unable to serialize data: %v", err)
		return
	}

	fmt.Println(string(res))

	deSer := new(SerTest)
	err = json.Unmarshal(res, deSer)

	if err != nil {
		t.Errorf("Unable to deserialize data: %v", err)
		return
	}

	if !deSer.Id1.IsEqual(t1.Id1) {
		t.Errorf("Deserialization of Id1 failed")
		return
	}

	if deSer.Id2 != t1.Id2 {
		t.Errorf("Deserialization of Id2 failed")
		return
	}

	if deSer.Egal.Compare(t1.Egal) != 0 {
		t.Errorf("Deserialization of time value failed")
		return
	}
}
