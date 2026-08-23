package model

import (
	"fmt"
	"strings"
	"time"
)

const (
	SampleDraft      = "draft"
	SampleMeasured   = "measured"
	SampleIdentified = "identified"
	SampleArchived   = "archived"
)

type Sample struct {
	ID          string
	SiteID      string
	Collector   string
	Condition   string
	Status      string
	Notes       string
	CollectedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s Sample) Validate() error {
	if strings.TrimSpace(s.ID) == "" || strings.TrimSpace(s.SiteID) == "" {
		return fmt.Errorf("sample identity is required")
	}
	if strings.TrimSpace(s.Collector) == "" {
		return fmt.Errorf("sample collector is required")
	}
	if s.CollectedAt.IsZero() {
		return fmt.Errorf("sample collection time is required")
	}
	return nil
}

func (s Sample) CanMeasure() bool  { return true }
func (s Sample) CanIdentify() bool { return s.Status == SampleMeasured }
func (s Sample) CanArchive() bool  { return s.Status == SampleIdentified }
