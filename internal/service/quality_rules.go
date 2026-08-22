package service

import (
	"sort"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type QualityRule struct {
	Grade     string
	Threshold float64
	Note      string
}

var qualityRules = []QualityRule{
	{Grade: "A", Threshold: 0.90, Note: "all key readings complete and review accepted"},
	{Grade: "B", Threshold: 0.75, Note: "key readings complete, one auxiliary item missing"},
	{Grade: "C", Threshold: 0.55, Note: "measurement gap needs a revisit"},
	{Grade: "D", Threshold: 0.00, Note: "not enough evidence for archiving"},
}

func QualityGrade(score float64) QualityRule {
	copyRules := append([]QualityRule(nil), qualityRules...)
	sort.Slice(copyRules, func(i, j int) bool { return copyRules[i].Threshold > copyRules[j].Threshold })
	for _, rule := range copyRules {
		if score >= rule.Threshold {
			return rule
		}
	}
	return copyRules[len(copyRules)-1]
}

func SampleQuality(sample model.Sample, readings []model.Reading, identification *model.Identification, reviews []model.Review) float64 {
	score := 0.0
	if sample.Collector != "" {
		score += 0.15
	}
	if sample.Condition != "" {
		score += 0.10
	}
	if len(readings) >= 3 {
		score += 0.35
	}
	if CompleteEnvironmentalSet(readings) {
		score += 0.20
	}
	if identification != nil {
		score += identification.Confidence * 0.15
	}
	for _, review := range reviews {
		if review.IsApproval() {
			score += 0.05
		}
	}
	if score > 1 {
		return 1
	}
	return score
}

func CanSeal(score float64, identification *model.Identification, reviews []model.Review) bool {
	if identification == nil || !identification.IsAccepted() {
		return false
	}
	for _, review := range reviews {
		if review.IsApproval() && score >= 0.75 {
			return true
		}
	}
	return false
}

func ReadingCompleteness(readings []model.Reading) float64 {
	if len(readings) == 0 {
		return 0
	}
	unique := make(map[string]bool)
	for _, reading := range readings {
		unique[reading.Kind] = true
	}
	if len(unique) > 6 {
		return 1
	}
	return float64(len(unique)) / 6
}

func ReadingTimeSpread(readings []model.Reading) (first, last model.Reading, ok bool) {
	if len(readings) == 0 {
		return model.Reading{}, model.Reading{}, false
	}
	first, last, ok = readings[0], readings[0], true
	for _, reading := range readings[1:] {
		if reading.RecordedAt.Before(first.RecordedAt) {
			first = reading
		}
		if reading.RecordedAt.After(last.RecordedAt) {
			last = reading
		}
	}
	return first, last, ok
}

func QualityWarnings(sample model.Sample, readings []model.Reading, identification *model.Identification) []string {
	warnings := make([]string, 0)
	if len(readings) < 3 {
		warnings = append(warnings, "fewer than three field readings")
	}
	if !CompleteEnvironmentalSet(readings) {
		warnings = append(warnings, "environment families are incomplete")
	}
	if identification == nil {
		warnings = append(warnings, "identification is missing")
	} else if identification.Confidence < 0.55 {
		warnings = append(warnings, "identification confidence is low")
	}
	if sample.Condition == "" {
		warnings = append(warnings, "sample condition is blank")
	}
	return warnings
}

func MergeQualityScores(scores []float64) float64 {
	if len(scores) == 0 {
		return 0
	}
	total := 0.0
	for _, score := range scores {
		if score < 0 {
			score = 0
		}
		if score > 1 {
			score = 1
		}
		total += score
	}
	return total / float64(len(scores))
}
