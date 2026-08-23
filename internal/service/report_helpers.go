package service

import (
	"sort"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func SortReportSamples(report *model.SiteReport) {
	sort.SliceStable(report.Samples, func(i, j int) bool {
		return report.Samples[i].Sample.CollectedAt.Before(report.Samples[j].Sample.CollectedAt)
	})
}

func ReportAge(now, created time.Time) time.Duration {
	if created.IsZero() || now.Before(created) {
		return 0
	}
	return now.Sub(created)
}
