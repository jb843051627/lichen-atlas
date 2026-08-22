package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/codec"
	"github.com/jb843051627/lichen-atlas/internal/model"
)

type ReportFilter struct {
	Region        string
	Status        string
	CollectedFrom time.Time
	CollectedTo   time.Time
	Limit         int
}

func (f ReportFilter) Normalize() ReportFilter {
	f.Region = strings.TrimSpace(strings.ToLower(f.Region))
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 100
	}
	return f
}

type ReportEnvelope struct {
	GeneratedAt time.Time
	Filter      ReportFilter
	Reports     []model.SiteReport
}

func BuildReportEnvelope(ctx context.Context, sites []model.Site, filter ReportFilter, build func(context.Context, string) (model.SiteReport, error)) (ReportEnvelope, error) {
	filter = filter.Normalize()
	envelope := ReportEnvelope{GeneratedAt: time.Now().UTC(), Filter: filter, Reports: make([]model.SiteReport, 0)}
	for _, site := range sites {
		if filter.Region != "" && strings.ToLower(site.Region) != filter.Region {
			continue
		}
		if err := ctx.Err(); err != nil {
			return ReportEnvelope{}, err
		}
		report, err := build(ctx, site.ID)
		if err != nil {
			return ReportEnvelope{}, fmt.Errorf("build site report %s: %w", site.ID, err)
		}
		if filter.Status != "" && report.Site.Status != filter.Status {
			continue
		}
		envelope.Reports = append(envelope.Reports, report)
		if len(envelope.Reports) >= filter.Limit {
			break
		}
	}
	sort.Slice(envelope.Reports, func(i, j int) bool { return envelope.Reports[i].Site.Name < envelope.Reports[j].Site.Name })
	return envelope, nil
}

func ReportJSON(envelope ReportEnvelope) ([]byte, error) {
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode report: %w", err)
	}
	return append(data, '\n'), nil
}

func ReportCSV(envelope ReportEnvelope) ([]byte, error) {
	var buffer bytes.Buffer
	for _, report := range envelope.Reports {
		if err := codec.WriteSiteReport(&buffer, report); err != nil {
			return nil, err
		}
	}
	return buffer.Bytes(), nil
}

func ReportStats(envelope ReportEnvelope) map[string]int {
	stats := map[string]int{"sites": 0, "samples": 0, "readings": 0, "open_samples": 0}
	for _, report := range envelope.Reports {
		stats["sites"]++
		stats["samples"] += len(report.Samples)
		stats["readings"] += report.TotalReadings
		stats["open_samples"] += report.OpenSamples
	}
	return stats
}

func ReportContainsUnreviewed(envelope ReportEnvelope) bool {
	for _, report := range envelope.Reports {
		if report.HasUnreviewedSamples() {
			return true
		}
	}
	return false
}
