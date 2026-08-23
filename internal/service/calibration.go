package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type Instrument struct {
	ID           string
	Serial       string
	Kind         string
	CalibratedAt time.Time
	ExpiresAt    time.Time
	Operator     string
}

func (i Instrument) Validate() error {
	if i.ID == "" || i.Serial == "" || i.Kind == "" {
		return fmt.Errorf("instrument identity is required")
	}
	if i.CalibratedAt.IsZero() || i.ExpiresAt.IsZero() {
		return fmt.Errorf("instrument calibration dates are required")
	}
	if !i.ExpiresAt.After(i.CalibratedAt) {
		return fmt.Errorf("instrument expiry must follow calibration")
	}
	return nil
}

func (i Instrument) IsValid(at time.Time) bool {
	return !i.CalibratedAt.IsZero() && !i.ExpiresAt.IsZero() && !at.Before(i.CalibratedAt) && at.Before(i.ExpiresAt)
}

func InstrumentForKind(instruments []Instrument, kind string, at time.Time) (Instrument, bool) {
	for _, instrument := range instruments {
		if strings.EqualFold(instrument.Kind, kind) && instrument.IsValid(at) {
			return instrument, true
		}
	}
	return Instrument{}, false
}

func ValidateReadingInstrument(reading model.Reading, instrument Instrument) error {
	if !instrument.IsValid(reading.RecordedAt) {
		return fmt.Errorf("instrument %s was not calibrated for reading time", instrument.Serial)
	}
	rule, ok := FindEnvironmentRule(reading.Kind)
	if !ok || !strings.EqualFold(rule.Unit, reading.Unit) {
		return fmt.Errorf("instrument reading unit is not accepted")
	}
	return nil
}

func CalibrationAge(at time.Time, instrument Instrument) time.Duration {
	if instrument.CalibratedAt.IsZero() || at.Before(instrument.CalibratedAt) {
		return 0
	}
	return at.Sub(instrument.CalibratedAt)
}
