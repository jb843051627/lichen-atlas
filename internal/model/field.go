package model

import (
	"fmt"
	"strings"
	"time"
)

type FieldVisit struct {
	ID          string
	SiteID      string
	Team        string
	Lead        string
	StartedAt   time.Time
	EndedAt     time.Time
	Weather     string
	AccessNotes string
}

func (v FieldVisit) Validate() error {
	if v.ID == "" || v.SiteID == "" || strings.TrimSpace(v.Team) == "" || strings.TrimSpace(v.Lead) == "" {
		return fmt.Errorf("field visit identity is required")
	}
	if v.StartedAt.IsZero() {
		return fmt.Errorf("field visit start is required")
	}
	if !v.EndedAt.IsZero() && v.EndedAt.Before(v.StartedAt) {
		return fmt.Errorf("field visit end precedes start")
	}
	return nil
}

func (v FieldVisit) IsOpen() bool { return v.EndedAt.IsZero() }

type SampleMark struct {
	Code      string
	SampleID  string
	VisitID   string
	Label     string
	Latitude  float64
	Longitude float64
	AccuracyM float64
	MarkedAt  time.Time
}

func (m SampleMark) Validate() error {
	if m.Code == "" || m.SampleID == "" || m.VisitID == "" {
		return fmt.Errorf("sample mark identity is required")
	}
	if m.Latitude < -90 || m.Latitude > 90 || m.Longitude < -180 || m.Longitude > 180 {
		return fmt.Errorf("sample mark coordinates are outside range")
	}
	if m.AccuracyM < 0 {
		return fmt.Errorf("sample mark accuracy cannot be negative")
	}
	return nil
}

type Substrate struct {
	Type       string
	Hardness   string
	Moisture   float64
	Chemistry  string
	Vegetation string
}

func (s Substrate) IsWet() bool  { return s.Moisture >= 60 }
func (s Substrate) IsRock() bool { return strings.EqualFold(s.Type, "rock") }

type SpecimenCondition struct {
	Thallus  string
	Color    string
	Damage   string
	Drying   string
	MassG    float64
	Observed time.Time
}

func (c SpecimenCondition) IsDamaged() bool {
	return c.Damage != "" && !strings.EqualFold(c.Damage, "none")
}

func (c SpecimenCondition) IsDry() bool { return strings.EqualFold(c.Drying, "complete") }

type Candidate struct {
	TaxonID    string
	Scientific string
	Confidence float64
	Evidence   []string
}

func (c Candidate) Validate() error {
	if c.TaxonID == "" || c.Scientific == "" {
		return fmt.Errorf("candidate identity is required")
	}
	if c.Confidence < 0 || c.Confidence > 1 {
		return fmt.Errorf("candidate confidence is outside range")
	}
	return nil
}

func SortCandidates(values []Candidate) []Candidate {
	result := append([]Candidate(nil), values...)
	for _, value := range result {
		value.Evidence = append([]string(nil), value.Evidence...)
	}
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Confidence > result[i].Confidence {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}
