package model

import (
	"fmt"
	"strings"
	"time"
)

type FieldTranscript struct {
	ID        string
	VisitID   string
	Author    string
	Text      string
	CreatedAt time.Time
	Tags      []string
}

func (t FieldTranscript) Validate() error {
	if t.ID == "" || t.VisitID == "" || strings.TrimSpace(t.Author) == "" {
		return fmt.Errorf("transcript identity is required")
	}
	if strings.TrimSpace(t.Text) == "" {
		return fmt.Errorf("transcript text is empty")
	}
	if t.CreatedAt.IsZero() {
		return fmt.Errorf("transcript time is required")
	}
	return nil
}

func (t FieldTranscript) HasTag(tag string) bool {
	for _, value := range t.Tags {
		if strings.EqualFold(value, tag) {
			return true
		}
	}
	return false
}
