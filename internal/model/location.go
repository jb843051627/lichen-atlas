package model

import "fmt"

type Location struct {
	Latitude  float64
	Longitude float64
	AccuracyM float64
}

func (l Location) Validate() error {
	if l.Latitude < -90 || l.Latitude > 90 || l.Longitude < -180 || l.Longitude > 180 {
		return fmt.Errorf("coordinates are outside range")
	}
	if l.AccuracyM < 0 {
		return fmt.Errorf("accuracy cannot be negative")
	}
	return nil
}
