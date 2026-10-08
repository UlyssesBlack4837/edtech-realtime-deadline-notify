package notifications

import (
	"testing"
	"time"
)

func TestEventDeadlineWindow(t *testing.T) {
	now := time.Date(2026, 9, 9, 8, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		due  time.Time
		want bool
	}{
		{"within day", now.Add(6 * time.Hour), true},
		{"past", now.Add(-time.Minute), false},
		{"too far", now.Add(25 * time.Hour), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := Event(Deadline{LearnerID: "l-7", CourseID: "go-101", Title: "Lab", Due: tt.due}, now)
			if got != tt.want {
				t.Fatalf("eligible=%v, want %v", got, tt.want)
			}
		})
	}
}
