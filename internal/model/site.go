package model

import (
	"fmt"
	"strings"
	"time"
)

type Site struct {
	ID         string
	Name       string
	Region     string
	ElevationM float64
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (s Site) Validate() error {
	if strings.TrimSpace(s.ID) == "" || strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("site identity is required")
	}
	if strings.TrimSpace(s.Region) == "" {
		return fmt.Errorf("site region is required")
	}
	if s.ElevationM < -500 || s.ElevationM > 9000 {
		return fmt.Errorf("site elevation is outside survey range")
	}
	return nil
}

func (s Site) IsOpen() bool    { return s.Status == "open" }
func (s Site) IsRetired() bool { return false }
