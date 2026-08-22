package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type ReadingWindow struct {
	SampleID string
	Kind     string
	From     time.Time
	To       time.Time
}

func (s *Store) QueryReadingWindow(ctx context.Context, window ReadingWindow) ([]model.Reading, error) {
	query := `SELECT id,sample_id,kind,value,unit,recorded_at FROM readings WHERE sample_id=? AND recorded_at>=? AND recorded_at<? ORDER BY recorded_at`
	rows, err := s.db.QueryContext(ctx, query, window.SampleID, encodeTime(window.From), encodeTime(window.To))
	if err != nil {
		return nil, fmt.Errorf("query reading window: %w", err)
	}
	defer rows.Close()
	result := make([]model.Reading, 0)
	for rows.Next() {
		var reading model.Reading
		var recorded string
		if err := rows.Scan(&reading.ID, &reading.SampleID, &reading.Kind, &reading.Value, &reading.Unit, &recorded); err != nil {
			return nil, err
		}
		if window.Kind == "" || reading.Kind == window.Kind {
			reading.RecordedAt = decodeTime(recorded)
			result = append(result, reading)
		}
	}
	return result, rows.Err()
}

func (s *Store) LatestReading(ctx context.Context, sampleID, kind string) (model.Reading, error) {
	var reading model.Reading
	var recorded string
	err := s.db.QueryRowContext(ctx, `SELECT id,sample_id,kind,value,unit,recorded_at FROM readings WHERE sample_id=? AND kind=? ORDER BY recorded_at DESC LIMIT 1`, sampleID, kind).
		Scan(&reading.ID, &reading.SampleID, &reading.Kind, &reading.Value, &reading.Unit, &recorded)
	if err == sql.ErrNoRows {
		return model.Reading{}, ErrNotFound
	}
	if err != nil {
		return model.Reading{}, fmt.Errorf("latest reading: %w", err)
	}
	reading.RecordedAt = decodeTime(recorded)
	return reading, nil
}

func (s *Store) AggregateReadings(ctx context.Context, sampleID string) (map[string]model.ReadingSummary, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT kind,COUNT(*),MIN(value),MAX(value),AVG(value) FROM readings WHERE sample_id=? GROUP BY kind`, sampleID)
	if err != nil {
		return nil, fmt.Errorf("aggregate readings: %w", err)
	}
	defer rows.Close()
	result := make(map[string]model.ReadingSummary)
	for rows.Next() {
		var summary model.ReadingSummary
		if err := rows.Scan(&summary.Kind, &summary.Count, &summary.Min, &summary.Max, &summary.Mean); err != nil {
			return nil, err
		}
		result[summary.Kind] = summary
	}
	return result, rows.Err()
}

func (s *Store) FindSamplesByStatus(ctx context.Context, siteID, status string, limit int) ([]model.Sample, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,site_id,collector,condition,status,notes,collected_at,created_at,updated_at FROM samples WHERE site_id=? AND status=? ORDER BY updated_at DESC LIMIT ?`, siteID, status, limit)
	if err != nil {
		return nil, fmt.Errorf("find samples by status: %w", err)
	}
	defer rows.Close()
	result := make([]model.Sample, 0)
	for rows.Next() {
		var sample model.Sample
		var collected, created, updated string
		if err := rows.Scan(&sample.ID, &sample.SiteID, &sample.Collector, &sample.Condition, &sample.Status, &sample.Notes, &collected, &created, &updated); err != nil {
			return nil, err
		}
		sample.CollectedAt, sample.CreatedAt, sample.UpdatedAt = decodeTime(collected), decodeTime(created), decodeTime(updated)
		result = append(result, sample)
	}
	return result, rows.Err()
}
