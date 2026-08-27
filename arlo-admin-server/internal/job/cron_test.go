package job

import (
	"testing"
	"time"
)

func TestParseCronAndMatch(t *testing.T) {
	spec, err := ParseCron("0 3 * * *")
	if err != nil {
		t.Fatal(err)
	}
	hit := time.Date(2026, 8, 4, 3, 0, 0, 0, time.Local)
	miss := time.Date(2026, 8, 4, 3, 1, 0, 0, time.Local)
	if !spec.Matches(hit) {
		t.Fatal("expected match at 03:00")
	}
	if spec.Matches(miss) {
		t.Fatal("should not match 03:01")
	}
	next := spec.Next(time.Date(2026, 8, 4, 2, 0, 0, 0, time.Local))
	if next.Hour() != 3 || next.Minute() != 0 {
		t.Fatalf("unexpected next: %v", next)
	}
}

func TestParseCronInvalid(t *testing.T) {
	if _, err := ParseCron("0 3 * *"); err == nil {
		t.Fatal("expected error")
	}
}
