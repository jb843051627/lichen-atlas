package model

import (
	"fmt"
	"strings"
	"time"
)

type Instrument struct {
	ID           string
	Serial       string
	Kind         string
	CalibratedAt time.Time
	ExpiresAt    time.Time
	Operator     string
	Status       string
}

func (i Instrument) Validate() error {
	if i.ID == "" || i.Serial == "" || strings.TrimSpace(i.Kind) == "" {
		return fmt.Errorf("instrument identity is required")
	}
	if i.CalibratedAt.IsZero() || i.ExpiresAt.IsZero() || !i.ExpiresAt.After(i.CalibratedAt) {
		return fmt.Errorf("instrument calibration window is invalid")
	}
	return nil
}

func (i Instrument) IsValid(at time.Time) bool {
	return !i.CalibratedAt.IsZero() && !i.ExpiresAt.IsZero() && !at.Before(i.CalibratedAt) && at.Before(i.ExpiresAt)
}

func (i Instrument) NeedsCalibration(at time.Time) bool { return !i.IsValid(at) }
