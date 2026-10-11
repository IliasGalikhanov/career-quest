package domain

import "errors"

var ErrIneligible = errors.New("ineligible")
var ErrActivityNotFound = errors.New("activity_not_found")
var ErrFutureSession = errors.New("future_session")
var ErrInvalidTransition = errors.New("invalid_transition")

func Start(b *Bundle, employeeID, eventID, recordID string) error {
	a, ok := Analyze(*b, employeeID)
	if !ok {
		return ErrActivityNotFound
	}
	for _, c := range a.Catalog {
		if c.Event.ID != eventID {
			continue
		}
		for _, reason := range c.Blocked {
			if reason != "no_gain" && reason != "no_goal" {
				return ErrIneligible
			}
		}
		b.History = append(b.History, Activity{ID: recordID, EmployeeID: employeeID, EventID: eventID, Date: b.Settings.Date, Status: "in_progress", AssignedBy: "self", SessionDate: c.SessionDate})
		return nil
	}
	return ErrActivityNotFound
}

func Finish(b *Bundle, recordID, status string) error {
	if status != "completed" && status != "dropped" {
		return ErrInvalidTransition
	}
	for i, h := range b.History {
		if h.ID != recordID {
			continue
		}
		if h.Status == status {
			return nil
		}
		if h.Status != "in_progress" {
			return ErrInvalidTransition
		}
		if h.Date > b.Settings.Date {
			return ErrFutureSession
		}
		if status == "completed" {
			if h.SessionDate > b.Settings.Date {
				return ErrFutureSession
			}
			h.Status = "completed"
			h.Completion = 100
			h.CompletedAt = b.Settings.Date
			h.Gains = []Gain{}
			for _, event := range b.Events {
				if event.ID == h.EventID {
					h.Gains = append(h.Gains, event.Gains...)
					break
				}
			}
		} else {
			h.Status = "dropped"
			h.Completion = max(5, h.Completion)
		}
		b.History[i] = h
		return nil
	}
	return ErrActivityNotFound
}
