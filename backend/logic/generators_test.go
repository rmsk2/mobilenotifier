package logic

import (
	"notifier/repo"
	"notifier/tools"
	"testing"
	"time"
)

func TestWeekly(t *testing.T) {
	tools.SetDefaultTZ()
	rem := repo.Reminder{}
	t1 := time.Date(2025, time.June, 3, 12, 0, 0, 0, tools.ClientTZ()).UTC()
	t2 := time.Date(2025, time.June, 1, 12, 0, 0, 0, tools.ClientTZ()).UTC()
	rem.Spec = t1

	t3 := weeklyRefTimeGen(&rem, t2).In(tools.ClientTZ())
	if (t3.Weekday() != t1.Weekday()) || (t3.Day() != t1.Day()) {
		t.Errorf("Test W1 failed")
	}

	t2 = time.Date(2025, time.June, 3, 12, 0, 0, 0, tools.ClientTZ()).UTC()
	t3 = weeklyRefTimeGen(&rem, t2).In(tools.ClientTZ())
	if (t3.Weekday() != t1.Weekday()) || (t3.Day() != 10) {
		t.Errorf("Test W2 failed")
	}

	t2 = time.Date(2025, time.June, 4, 12, 0, 0, 0, tools.ClientTZ()).UTC()
	t3 = weeklyRefTimeGen(&rem, t2).In(tools.ClientTZ())
	if (t3.Weekday() != t1.Weekday()) || (t3.Day() != 10) {
		t.Errorf("Test W3 failed")
	}
}

func TestYearly(t *testing.T) {
	tools.SetDefaultTZ()
	rem := repo.Reminder{}
	t1 := time.Date(2025, time.June, 3, 12, 0, 0, 0, tools.ClientTZ()).UTC()
	t2 := time.Date(2025, time.June, 1, 12, 14, 5, 0, tools.ClientTZ()).UTC()
	rem.Spec = t1

	t3 := anniversaryRefTimeGen(&rem, t2).In(tools.ClientTZ())
	if (t3.Year() != t1.Year()) || (t3.Month() != t1.Month()) || (t3.Day() != t1.Day()) {
		t.Errorf("Test Y1 failed: %v", t3)
	}

	t2 = time.Date(2025, time.June, 3, 0, 0, 0, 0, tools.ClientTZ()).UTC()
	t3 = anniversaryRefTimeGen(&rem, t2).In(tools.ClientTZ())
	if (t3.Year() != t1.Year()+1) || (t3.Month() != t1.Month()) || (t3.Day() != t1.Day()) {
		t.Errorf("Test Y5 failed: %v", t3)
	}

	t2 = time.Date(2025, time.June, 3, 0, 0, 0, 1, tools.ClientTZ()).UTC()
	t3 = anniversaryRefTimeGen(&rem, t2).In(tools.ClientTZ())
	if (t3.Year() != t1.Year()+1) || (t3.Month() != t1.Month()) || (t3.Day() != t1.Day()) {
		t.Errorf("Test Y4 failed: %v", t3)
	}

	t2 = time.Date(2025, time.June, 3, 12, 17, 34, 0, tools.ClientTZ()).UTC()
	t3 = anniversaryRefTimeGen(&rem, t2).In(tools.ClientTZ())
	if (t3.Year() != t1.Year()+1) || (t3.Month() != t1.Month()) || (t3.Day() != t1.Day()) {
		t.Errorf("Test Y2 failed")
	}

	t2 = time.Date(2025, time.June, 4, 12, 8, 25, 0, tools.ClientTZ()).UTC()
	t3 = anniversaryRefTimeGen(&rem, t2).In(tools.ClientTZ())
	if (t3.Year() != t1.Year()+1) || (t3.Month() != t1.Month()) || (t3.Day() != t1.Day()) {
		t.Errorf("Test Y3 failed")
	}
}
