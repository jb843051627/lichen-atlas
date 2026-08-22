package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type ReportSection struct {
	Name     string
	Priority int
	Lines    []string
}

func BuildReportSections(report model.SiteReport) []ReportSection {
	sections := []ReportSection{
		{Name: "site", Priority: 1, Lines: []string{report.Site.ID, report.Site.Name, report.Site.Region}},
		{Name: "coverage", Priority: 2, Lines: []string{fmt.Sprintf("samples=%d", len(report.Samples)), fmt.Sprintf("readings=%d", report.TotalReadings)}},
	}
	for _, sample := range report.Samples {
		lines := []string{sample.Sample.ID, sample.Sample.Status, sample.Sample.Collector}
		if sample.Identification != nil {
			lines = append(lines, sample.Identification.TaxonID, fmt.Sprintf("confidence=%.3f", sample.Identification.Confidence))
		}
		sections = append(sections, ReportSection{Name: "sample:" + sample.Sample.ID, Priority: 10, Lines: lines})
	}
	sort.SliceStable(sections, func(i, j int) bool { return sections[i].Priority < sections[j].Priority })
	return sections
}

func FlattenSection(section ReportSection) string {
	lines := make([]string, 0, len(section.Lines))
	for _, line := range section.Lines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, " | ")
}

func FilterSections(sections []ReportSection, prefix string) []ReportSection {
	result := make([]ReportSection, 0)
	for _, section := range sections {
		if prefix == "" || strings.HasPrefix(section.Name, prefix) {
			section.Lines = append([]string(nil), section.Lines...)
			result = append(result, section)
		}
	}
	return result
}

func SectionIndex(sections []ReportSection) map[string]int {
	result := make(map[string]int, len(sections))
	for index, section := range sections {
		result[section.Name] = index
	}
	return result
}
