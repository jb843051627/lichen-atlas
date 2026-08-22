package model

import "time"

type Event struct {
	ID        string
	SampleID  string
	EventType string
	Payload   string
	CreatedAt time.Time
}

func (e Event) IsStateChange() bool {
	switch e.EventType {
	case "sample.created", "sample.measured", "sample.identified", "sample.archived":
		return true
	default:
		return false
	}
}
