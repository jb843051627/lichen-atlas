package model

import "time"

type Task struct {
	ID          string
	SampleID    string
	Kind        string
	State       string
	Attempts    int
	AvailableAt time.Time
	ClaimedAt   time.Time
	FinishedAt  time.Time
	LastError   string
}

func (t Task) IsReady(now time.Time) bool {
	return t.State == "queued" && !t.AvailableAt.After(now)
}

func (t Task) IsTerminal() bool { return t.State == "done" || t.State == "cancelled" }
