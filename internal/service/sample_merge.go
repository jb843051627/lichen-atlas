package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func MergeReadings(left, right []model.Reading) ([]model.Reading, error) {
	result := make([]model.Reading, 0, len(left)+len(right))
	seen := make(map[string]model.Reading)
	for _, reading := range append(append([]model.Reading{}, left...), right...) {
		if reading.ID == "" {
			return nil, fmt.Errorf("reading id is empty")
		}
		if old, ok := seen[reading.ID]; ok {
			if old.Value != reading.Value || old.RecordedAt != reading.RecordedAt {
				return nil, fmt.Errorf("reading %s has conflicting values", reading.ID)
			}
			continue
		}
		seen[reading.ID] = reading
		result = append(result, reading)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].RecordedAt.Before(result[j].RecordedAt) })
	return result, nil
}

func MergeSampleNotes(sample model.Sample, values []string) model.Sample {
	parts := make([]string, 0, len(values)+1)
	if strings.TrimSpace(sample.Notes) != "" {
		parts = append(parts, sample.Notes)
	}
	parts = append(parts, values...)
	sample.Notes = MergeNotes(parts)
	return sample
}

func CloneReport(report model.SampleReport) model.SampleReport {
	result := report
	result.Readings = append([]model.Reading(nil), report.Readings...)
	result.Reviews = append([]model.Review(nil), report.Reviews...)
	result.ReadingSummary = append([]model.ReadingSummary(nil), report.ReadingSummary...)
	return result
}
