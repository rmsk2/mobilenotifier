package logic

import (
	"notifier/repo"
	"testing"
	"time"
)

func TestWeekly(t *testing.T) {
	rem := repo.Reminder{}
	t1 := time.Date(2025, time.June, 3, 12, 0, 0, 0, time.Local)
	t2 := time.Date(2025, time.June, 1, 12, 0, 0, 0, time.Local)
	rem.Spec = t1

	t3 := weeklyRefTimeGen(&rem, t2)
	if (t3.Weekday() != t1.Weekday()) || (t3.Day() != t1.Day()) {
		t.Errorf("Test W1 failed")
	}

	t2 = time.Date(2025, time.June, 3, 12, 0, 0, 0, time.Local)
	t3 = weeklyRefTimeGen(&rem, t2)
	if (t3.Weekday() != t1.Weekday()) || (t3.Day() != 10) {
		t.Errorf("Test W2 failed")
	}

	t2 = time.Date(2025, time.June, 4, 12, 0, 0, 0, time.Local)
	t3 = weeklyRefTimeGen(&rem, t2)
	if (t3.Weekday() != t1.Weekday()) || (t3.Day() != 10) {
		t.Errorf("Test W3 failed")
	}
}

func TestYearly(t *testing.T) {
	rem := repo.Reminder{}
	t1 := time.Date(2025, time.June, 3, 12, 0, 0, 0, time.Local)
	t2 := time.Date(2025, time.June, 1, 12, 14, 5, 0, time.Local)
	rem.Spec = t1

	t3 := anniversaryRefTimeGen(&rem, t2)
	if (t3.Year() != t1.Year()) || (t3.Month() != t1.Month()) || (t3.Day() != t1.Day()) {
		t.Errorf("Test Y1 failed: %v", t3)
	}

	t2 = time.Date(2025, time.June, 3, 12, 17, 34, 0, time.Local)
	t3 = anniversaryRefTimeGen(&rem, t2)
	if (t3.Year() != t1.Year()+1) || (t3.Month() != t1.Month()) || (t3.Day() != t1.Day()) {
		t.Errorf("Test Y2 failed")
	}

	t2 = time.Date(2025, time.June, 4, 12, 8, 25, 0, time.Local)
	t3 = anniversaryRefTimeGen(&rem, t2)
	if (t3.Year() != t1.Year()+1) || (t3.Month() != t1.Month()) || (t3.Day() != t1.Day()) {
		t.Errorf("Test Y3 failed")
	}
}
