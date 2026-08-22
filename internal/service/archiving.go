package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type ArchiveBatch struct {
	ID        string
	Operator  string
	BoxCode   string
	Samples   []model.ArchiveRecord
	StartedAt time.Time
}

func (b ArchiveBatch) Validate() error {
	if b.ID == "" || strings.TrimSpace(b.Operator) == "" || strings.TrimSpace(b.BoxCode) == "" {
		return fmt.Errorf("archive batch identity is required")
	}
	if len(b.Samples) == 0 {
		return fmt.Errorf("archive batch is empty")
	}
	seen := make(map[string]bool)
	for _, item := range b.Samples {
		if item.SampleID == "" || seen[item.SampleID] {
			return fmt.Errorf("archive sample is repeated or empty")
		}
		seen[item.SampleID] = true
	}
	return nil
}

func SortArchiveRecords(values []model.ArchiveRecord) []model.ArchiveRecord {
	result := append([]model.ArchiveRecord(nil), values...)
	sort.SliceStable(result, func(i, j int) bool { return result[i].SampleID < result[j].SampleID })
	return result
}

func ArchiveBoxLabel(batch ArchiveBatch) string {
	if batch.BoxCode == "" {
		return "unassigned"
	}
	return strings.ToUpper(strings.TrimSpace(batch.BoxCode))
}

func ArchiveReady(sample model.Sample, identification *model.Identification, reviews []model.Review, readings []model.Reading) (float64, bool) {
	score := SampleQuality(sample, readings, identification, reviews)
	return score, CanSeal(score, identification, reviews)
}

func BuildArchiveBatch(ctx context.Context, operator, box string, samples []model.Sample, load func(context.Context, string) (model.SampleReport, error), now time.Time) (ArchiveBatch, error) {
	batch := ArchiveBatch{ID: "archive-" + now.Format("20060102150405"), Operator: operator, BoxCode: box, StartedAt: now}
	for _, sample := range samples {
		if err := ctx.Err(); err != nil {
			return ArchiveBatch{}, err
		}
		report, err := load(ctx, sample.ID)
		if err != nil {
			return ArchiveBatch{}, fmt.Errorf("load sample %s for archive: %w", sample.ID, err)
		}
		score, ok := ArchiveReady(report.Sample, report.Identification, report.Reviews, report.Readings)
		if !ok {
			return ArchiveBatch{}, fmt.Errorf("sample %s is below archive quality threshold %.2f", sample.ID, score)
		}
		batch.Samples = append(batch.Samples, model.ArchiveRecord{ID: batch.ID + "-" + sample.ID, SampleID: sample.ID, BoxCode: box, SealedBy: operator, SealState: "sealed", SealedAt: now})
	}
	return batch, batch.Validate()
}

func ArchiveSummary(records []model.ArchiveRecord) map[string]int {
	result := make(map[string]int)
	for _, record := range records {
		result[record.SealState]++
	}
	return result
}

func IsOperatorAllowed(operator string, sample model.Sample, reviews []model.Review) bool {
	operator = strings.TrimSpace(operator)
	if operator == "" {
		return false
	}
	for _, review := range reviews {
		if review.IsApproval() && review.Reviewer == operator {
			return false
		}
	}
	return sample.Collector != operator
}

func SealedBefore(records []model.ArchiveRecord, cutoff time.Time) []model.ArchiveRecord {
	result := make([]model.ArchiveRecord, 0)
	for _, record := range records {
		if !record.SealedAt.IsZero() && record.SealedAt.Before(cutoff) {
			result = append(result, record)
		}
	}
	return SortArchiveRecords(result)
}
