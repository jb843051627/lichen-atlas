package model

import "time"

type ReadingSummary struct {
	Kind  string  `json:"kind"`
	Count int     `json:"count"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Mean  float64 `json:"mean"`
}

type SampleReport struct {
	Sample         Sample
	Readings       []Reading
	ReadingSummary []ReadingSummary
	Identification *Identification
	Reviews        []Review
	Archive        *ArchiveRecord
}

type SiteReport struct {
	Site          Site
	Samples       []SampleReport
	GeneratedAt   time.Time
	TotalReadings int
	OpenSamples   int
}

func (r SiteReport) HasUnreviewedSamples() bool {
	for _, sample := range r.Samples {
		if sample.Identification != nil && len(sample.Reviews) == 0 {
			return true
		}
	}
	return false
}
