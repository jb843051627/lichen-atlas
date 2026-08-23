package service

import (
	"strings"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func FilterReadings(values []model.Reading, kind string) []model.Reading {
	result := make([]model.Reading, 0, len(values))
	for _, value := range values {
		if kind == "" || strings.EqualFold(value.Kind, kind) {
			result = append(result, value)
		}
	}
	return result
}

func LatestReading(values []model.Reading, kind string) (model.Reading, bool) {
	var result model.Reading
	found := false
	for _, value := range values {
		if (kind == "" || value.Kind == kind) && (!found || result.RecordedAt.Before(value.RecordedAt)) {
			result, found = value, true
		}
	}
	return result, found
}
