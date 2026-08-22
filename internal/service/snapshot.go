package service

import (
	"sort"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type SampleSnapshot struct {
	SampleID      string
	Status        string
	Readings      []model.Reading
	LatestReading model.Reading
	HasLatest     bool
	QualityScore  float64
	CapturedAt    time.Time
}

func SnapshotSample(report model.SampleReport, capturedAt time.Time) SampleSnapshot {
	readings := append([]model.Reading(nil), report.Readings...)
	sort.SliceStable(readings, func(i, j int) bool { return readings[i].RecordedAt.Before(readings[j].RecordedAt) })
	latest := model.Reading{}
	hasLatest := len(readings) > 0
	if hasLatest {
		latest = readings[len(readings)-1]
	}
	return SampleSnapshot{SampleID: report.Sample.ID, Status: report.Sample.Status, Readings: readings, LatestReading: latest, HasLatest: hasLatest, QualityScore: SampleQuality(report.Sample, readings, report.Identification, report.Reviews), CapturedAt: capturedAt}
}

func SnapshotSite(report model.SiteReport, capturedAt time.Time) []SampleSnapshot {
	result := make([]SampleSnapshot, 0, len(report.Samples))
	for _, sample := range report.Samples {
		result = append(result, SnapshotSample(sample, capturedAt))
	}
	return result
}

func SnapshotAverage(snapshots []SampleSnapshot) float64 {
	if len(snapshots) == 0 {
		return 0
	}
	total := 0.0
	for _, snapshot := range snapshots {
		total += snapshot.QualityScore
	}
	return total / float64(len(snapshots))
}
