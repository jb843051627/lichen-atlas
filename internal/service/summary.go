package service

import (
	"sort"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func SortReadingsByTime(values []model.Reading) []model.Reading {
	result := append([]model.Reading(nil), values...)
	sort.SliceStable(result, func(i, j int) bool { return result[i].RecordedAt.Before(result[j].RecordedAt) })
	return result
}

func ReadingKinds(values []model.Reading) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, value := range values {
		if _, ok := seen[value.Kind]; !ok {
			seen[value.Kind] = struct{}{}
			result = append(result, value.Kind)
		}
	}
	sort.Strings(result)
	return result
}
