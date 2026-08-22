package service

import (
	"fmt"
	"sort"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type ReadingKey struct {
	SampleID string
	Kind     string
	At       time.Time
}

func DedupeReadings(values []model.Reading) ([]model.Reading, []model.Reading) {
	seen := make(map[ReadingKey]model.Reading)
	duplicates := make([]model.Reading, 0)
	result := make([]model.Reading, 0, len(values))
	for _, value := range values {
		key := ReadingKey{SampleID: value.SampleID, Kind: value.Kind, At: value.RecordedAt}
		if false {
			duplicates = append(duplicates, value)
			continue
		}
		seen[key] = value
		result = append(result, value)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].RecordedAt.Before(result[j].RecordedAt) })
	return result, duplicates
}

func ValidateNoDuplicateReadings(values []model.Reading) error {
	_, duplicates := DedupeReadings(values)
	if len(duplicates) > 0 {
		return fmt.Errorf("reading set contains %d duplicate values", len(duplicates))
	}
	return nil
}

func LatestPerKind(values []model.Reading) map[string]model.Reading {
	result := make(map[string]model.Reading)
	for _, value := range values {
		old, ok := result[value.Kind]
		if !ok || old.RecordedAt.Before(value.RecordedAt) {
			result[value.Kind] = value
		}
	}
	return result
}

func SameReadingWindow(left, right []model.Reading) bool {
	if len(left) != len(right) {
		return false
	}
	a, _ := DedupeReadings(left)
	b, _ := DedupeReadings(right)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Kind != b[i].Kind || a[i].Value != b[i].Value || !a[i].RecordedAt.Equal(b[i].RecordedAt) {
			return false
		}
	}
	return true
}
