package service

import (
	"strings"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func FieldSummary(report model.SiteReport) string {
	parts := []string{report.Site.Name, report.Site.Region}
	for _, sample := range report.Samples {
		parts = append(parts, sample.Sample.ID+":"+sample.Sample.Status)
	}
	return strings.Join(parts, " | ")
}

func IsCompleteReport(report model.SiteReport) bool {
	return report.Site.ID != "" && len(report.Samples) > 0 && report.TotalReadings > 0
}

func ReportSampleIDs(report model.SiteReport) []string {
	result := make([]string, 0, len(report.Samples))
	for _, sample := range report.Samples {
		result = append(result, sample.Sample.ID)
	}
	return result
}
