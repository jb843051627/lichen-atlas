package service

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

func NormalizeScientificName(value string) string {
	words := strings.Fields(strings.TrimSpace(value))
	for i, word := range words {
		if word == "" {
			continue
		}
		runes := []rune(strings.ToLower(word))
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func NormalizeUnit(unit string) string {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "c", "°c", "degc":
		return "C"
	case "percent", "%", "pct":
		return "%"
	case "lux":
		return "lux"
	default:
		return strings.TrimSpace(unit)
	}
}

func ClampScore(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func ValidatePageLimit(limit int) error {
	if limit < 0 || limit > 500 {
		return fmt.Errorf("page limit must be between zero and 500")
	}
	return nil
}

func MergeNotes(values []string) string {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool)
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return strings.Join(result, "; ")
}
