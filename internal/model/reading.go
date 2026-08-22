package model

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Reading struct {
	ID         string
	SampleID   string
	Kind       string
	Value      float64
	Unit       string
	RecordedAt time.Time
}

func (r Reading) Validate() error {
	if strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.SampleID) == "" {
		return fmt.Errorf("reading identity is required")
	}
	if strings.TrimSpace(r.Kind) == "" || strings.TrimSpace(r.Unit) == "" {
		return fmt.Errorf("reading kind and unit are required")
	}
	if math.IsNaN(r.Value) || math.IsInf(r.Value, 0) {
		return fmt.Errorf("reading value is not finite")
	}
	if r.RecordedAt.IsZero() {
		return fmt.Errorf("reading timestamp is required")
	}
	return nil
}

func (r Reading) IsEnvironmental() bool {
	switch strings.ToLower(r.Kind) {
	case "temperature", "humidity", "light", "substrate":
		return true
	default:
		return false
	}
}
