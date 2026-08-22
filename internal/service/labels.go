package service

import (
	"sort"
	"strings"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func SampleLabels(sample model.Sample, report model.SampleReport) []string {
	labels := make([]string, 0, 8)
	if sample.Status != "" {
		labels = append(labels, "status:"+sample.Status)
	}
	if sample.Condition != "" {
		labels = append(labels, "condition:"+strings.ToLower(sample.Condition))
	}
	if len(report.Readings) >= 3 {
		labels = append(labels, "measured")
	}
	if CompleteEnvironmentalSet(report.Readings) {
		labels = append(labels, "environment-complete")
	}
	if report.Identification != nil {
		labels = append(labels, "identified")
	}
	if len(report.Reviews) > 0 {
		labels = append(labels, "reviewed")
	}
	if report.Archive != nil {
		labels = append(labels, "sealed")
	}
	sort.Strings(labels)
	return labels
}

func HasLabel(labels []string, expected string) bool {
	for _, label := range labels {
		if label == expected {
			return true
		}
	}
	return false
}

func AddLabel(labels []string, value string) []string {
	if strings.TrimSpace(value) == "" || HasLabel(labels, value) {
		return append([]string(nil), labels...)
	}
	result := append(append([]string(nil), labels...), value)
	sort.Strings(result)
	return result
}

func RemoveLabel(labels []string, value string) []string {
	result := make([]string, 0, len(labels))
	for _, label := range labels {
		if label != value {
			result = append(result, label)
		}
	}
	return result
}
