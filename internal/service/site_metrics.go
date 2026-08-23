package service

import (
	"math"
	"sort"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type SiteMetrics struct {
	SiteID             string
	SampleCount        int
	ArchivedCount      int
	MeasuredCount      int
	IdentificationRate float64
	ArchiveRate        float64
	ReadingMean        float64
	LatestCollection   time.Time
}

func BuildSiteMetrics(report model.SiteReport) SiteMetrics {
	metrics := SiteMetrics{SiteID: report.Site.ID, SampleCount: len(report.Samples), LatestCollection: time.Time{}}
	totalReadings := 0
	totalValue := 0.0
	for _, sample := range report.Samples {
		switch sample.Sample.Status {
		case model.SampleArchived:
			metrics.ArchivedCount++
		case model.SampleMeasured:
			metrics.MeasuredCount++
		}
		if sample.Identification != nil {
			metrics.IdentificationRate++
		}
		if sample.Sample.CollectedAt.After(metrics.LatestCollection) {
			metrics.LatestCollection = sample.Sample.CollectedAt
		}
		for _, reading := range sample.Readings {
			totalReadings++
			totalValue += reading.Value
		}
	}
	if metrics.SampleCount > 0 {
		metrics.IdentificationRate /= float64(metrics.SampleCount)
		metrics.ArchiveRate = float64(metrics.ArchivedCount) / float64(metrics.SampleCount)
	}
	if totalReadings > 0 {
		metrics.ReadingMean = totalValue / float64(totalReadings)
	}
	return metrics
}

func RankSitesByArchiveRate(metrics []SiteMetrics) []SiteMetrics {
	result := append([]SiteMetrics(nil), metrics...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].ArchiveRate == result[j].ArchiveRate {
			return result[i].SiteID < result[j].SiteID
		}
		return result[i].ArchiveRate > result[j].ArchiveRate
	})
	return result
}

func MetricsHealth(metrics SiteMetrics) string {
	if metrics.SampleCount == 0 {
		return "empty"
	}
	if metrics.ArchiveRate >= 0.75 && metrics.IdentificationRate >= 0.75 {
		return "stable"
	}
	if metrics.IdentificationRate >= 0.5 {
		return "review"
	}
	return "attention"
}

func SafeRate(numerator, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	value := float64(numerator) / float64(denominator)
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return value
}
