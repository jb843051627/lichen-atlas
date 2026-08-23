package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type ExportPolicy struct {
	Format          string
	IncludeRaw      bool
	IncludeLocation bool
	IncludeNotes    bool
	Timezone        *time.Location
	MaxSamples      int
}

func (p ExportPolicy) Validate() error {
	switch strings.ToLower(p.Format) {
	case "json", "csv":
	default:
		return fmt.Errorf("unsupported export format %q", p.Format)
	}
	if p.MaxSamples < 0 || p.MaxSamples > 500 {
		return fmt.Errorf("export sample limit is outside range")
	}
	if p.Timezone == nil {
		return fmt.Errorf("export timezone is required")
	}
	return nil
}

func (p ExportPolicy) Normalize() ExportPolicy {
	p.Format = strings.ToLower(strings.TrimSpace(p.Format))
	if p.MaxSamples == 0 {
		p.MaxSamples = 100
	}
	if p.Timezone == nil {
		p.Timezone = time.UTC
	}
	return p
}

func ExportSampleAllowed(sample model.Sample, policy ExportPolicy) bool {
	if policy.MaxSamples <= 0 {
		return false
	}
	if sample.ID == "" || sample.SiteID == "" {
		return false
	}
	return true
}

func ExportTimestamp(value time.Time, policy ExportPolicy) string {
	if value.IsZero() {
		return ""
	}
	location := policy.Timezone
	if location == nil {
		location = time.UTC
	}
	return value.UTC().Format(time.RFC3339)
}

func ExportStatus(sample model.Sample) string {
	switch sample.Status {
	case model.SampleDraft:
		return "field"
	case model.SampleMeasured:
		return "measured"
	case model.SampleIdentified:
		return "identified"
	case model.SampleArchived:
		return "archived"
	default:
		return "unknown"
	}
}

func ExportReadingRows(report model.SampleReport, policy ExportPolicy) [][]string {
	rows := make([][]string, 0, len(report.Readings))
	for _, reading := range report.Readings {
		rows = append(rows, []string{report.Sample.ID, ExportStatus(report.Sample), reading.Kind, fmt.Sprintf("%.6f", reading.Value), reading.Unit, ExportTimestamp(reading.RecordedAt, policy)})
	}
	return rows
}

func ExportSummary(report model.SiteReport, policy ExportPolicy) map[string]string {
	policy = policy.Normalize()
	return map[string]string{
		"site_id":      report.Site.ID,
		"site_name":    report.Site.Name,
		"region":       report.Site.Region,
		"generated_at": report.GeneratedAt.UTC().Format(time.RFC3339),
		"samples":      fmt.Sprintf("%d", len(report.Samples)),
		"readings":     fmt.Sprintf("%d", report.TotalReadings),
		"open_samples": fmt.Sprintf("%d", report.OpenSamples),
	}
}

func RedactReport(report model.SampleReport, policy ExportPolicy) model.SampleReport {
	result := CloneReport(report)
	if !policy.IncludeNotes {
		result.Sample.Notes = ""
	}
	if !policy.IncludeRaw {
		for index := range result.Readings {
			result.Readings[index].ID = ""
		}
	}
	return result
}

func ExportBatch(reports []model.SampleReport, policy ExportPolicy) []model.SampleReport {
	policy = policy.Normalize()
	result := make([]model.SampleReport, 0, len(reports))
	for _, report := range reports {
		if len(result) >= policy.MaxSamples {
			break
		}
		if ExportSampleAllowed(report.Sample, policy) {
			result = append(result, RedactReport(report, policy))
		}
	}
	return result
}
