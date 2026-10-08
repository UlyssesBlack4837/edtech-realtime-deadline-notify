package notifications

import "time"

type Deadline struct {
	LearnerID string
	CourseID  string
	Title     string
	Due       time.Time
}

// Event returns the business event when a learner is due soon.
func Event(d Deadline, now time.Time) (string, bool) {
	if d.LearnerID == "" || d.CourseID == "" || d.Title == "" {
		return "", false
	}
	remaining := d.Due.Sub(now)
	if remaining < 0 || remaining > 24*time.Hour {
		return "", false
	}
	return "deadline.reminder", true
}
