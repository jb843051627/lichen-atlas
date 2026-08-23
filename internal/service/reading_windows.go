package service

import (
	"fmt"
	"sort"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type ReadingWindowStats struct {
	Kind     string
	Count    int
	First    time.Time
	Last     time.Time
	GapCount int
	MeanGap  time.Duration
	Min      float64
	Max      float64
	Mean     float64
}

func StatsForReadingWindow(values []model.Reading, expectedGap time.Duration) ReadingWindowStats {
	if len(values) == 0 {
		return ReadingWindowStats{}
	}
	ordered := SortReadingsByTime(values)
	result := ReadingWindowStats{Kind: ordered[0].Kind, Count: len(ordered), First: ordered[0].RecordedAt, Last: ordered[len(ordered)-1].RecordedAt, Min: ordered[0].Value, Max: ordered[0].Value}
	total, gapTotal := 0.0, time.Duration(0)
	for i, value := range ordered {
		if value.Value < result.Min {
			result.Min = value.Value
		}
		if value.Value > result.Max {
			result.Max = value.Value
		}
		total += value.Value
		if i > 0 {
			gap := value.RecordedAt.Sub(ordered[i-1].RecordedAt)
			if expectedGap > 0 && gap > expectedGap {
				result.GapCount++
			}
			gapTotal += gap
		}
	}
	result.Mean = total / float64(len(ordered))
	if len(ordered) > 1 {
		result.MeanGap = gapTotal / time.Duration(len(ordered)-1)
	}
	return result
}

func GroupReadingWindows(values []model.Reading, expectedGap time.Duration) map[string]ReadingWindowStats {
	groups := make(map[string][]model.Reading)
	for _, value := range values {
		groups[value.Kind] = append(groups[value.Kind], value)
	}
	result := make(map[string]ReadingWindowStats, len(groups))
	for kind, group := range groups {
		stats := StatsForReadingWindow(group, expectedGap)
		stats.Kind = kind
		result[kind] = stats
	}
	return result
}

func DetectOutliers(values []model.Reading, lower, upper float64) []model.Reading {
	result := make([]model.Reading, 0)
	for _, value := range values {
		if value.Value < lower || value.Value > upper {
			result = append(result, value)
		}
	}
	return SortReadingsByTime(result)
}

func ValidateReadingSequence(values []model.Reading) error {
	if len(values) == 0 {
		return fmt.Errorf("reading sequence is empty")
	}
	ordered := append([]model.Reading(nil), values...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].RecordedAt.Before(ordered[j].RecordedAt) })
	for i := 1; i < len(ordered); i++ {
		if ordered[i].SampleID != ordered[0].SampleID {
			return fmt.Errorf("reading sequence contains multiple samples")
		}
		if ordered[i].RecordedAt.Equal(ordered[i-1].RecordedAt) && ordered[i].Kind == ordered[i-1].Kind {
			return fmt.Errorf("duplicate reading timestamp for %s", ordered[i].Kind)
		}
	}
	return nil
}
