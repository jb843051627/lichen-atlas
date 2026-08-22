package model

import (
	"fmt"
	"time"
)

type Identification struct {
	ID         string
	SampleID   string
	TaxonID    string
	Reviewer   string
	Confidence float64
	Status     string
	Comment    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (i Identification) Validate() error {
	if i.ID == "" || i.SampleID == "" || i.TaxonID == "" || i.Reviewer == "" {
		return fmt.Errorf("identification fields are required")
	}
	if i.Confidence < 0 || i.Confidence > 1 {
		return fmt.Errorf("identification confidence is outside range")
	}
	return nil
}

func (i Identification) IsAccepted() bool { return i.Status == "accepted" }
