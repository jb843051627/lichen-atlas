package model

import "time"

type FieldNote struct {
	ID        string
	SampleID  string
	Author    string
	Text      string
	CreatedAt time.Time
}

func (n FieldNote) IsEmpty() bool { return n.Text == "" }
