package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type SiteDigest struct {
	SiteID       string
	SiteName     string
	Region       string
	SampleCount  int
	ReadingCount int
	LatestSample time.Time
}

func (s *Store) SiteDigest(ctx context.Context, siteID string) (SiteDigest, error) {
	var digest SiteDigest
	var latest string
	err := s.db.QueryRowContext(ctx, `SELECT s.id,s.name,s.region,COUNT(DISTINCT sm.id),COUNT(r.id),COALESCE(MAX(sm.collected_at),'') FROM sites s LEFT JOIN samples sm ON sm.site_id=s.id LEFT JOIN readings r ON r.sample_id=sm.id WHERE s.id=? GROUP BY s.id,s.name,s.region`, siteID).
		Scan(&digest.SiteID, &digest.SiteName, &digest.Region, &digest.SampleCount, &digest.ReadingCount, &latest)
	if err != nil {
		return SiteDigest{}, fmt.Errorf("site digest: %w", err)
	}
	digest.LatestSample = decodeTime(latest)
	return digest, nil
}

func (s *Store) ListSiteDigests(ctx context.Context, region string, limit int) ([]SiteDigest, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := `SELECT s.id,s.name,s.region,COUNT(DISTINCT sm.id),COUNT(r.id),COALESCE(MAX(sm.collected_at),'') FROM sites s LEFT JOIN samples sm ON sm.site_id=s.id LEFT JOIN readings r ON r.sample_id=sm.id`
	args := []any{}
	if region != "" {
		query += ` WHERE s.region=?`
		args = append(args, region)
	}
	query += ` GROUP BY s.id,s.name,s.region ORDER BY s.name LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list site digests: %w", err)
	}
	defer rows.Close()
	result := make([]SiteDigest, 0)
	for rows.Next() {
		var digest SiteDigest
		var latest string
		if err := rows.Scan(&digest.SiteID, &digest.SiteName, &digest.Region, &digest.SampleCount, &digest.ReadingCount, &latest); err != nil {
			return nil, err
		}
		digest.LatestSample = decodeTime(latest)
		result = append(result, digest)
	}
	return result, rows.Err()
}

func (s *Store) SamplesCollectedBetween(ctx context.Context, siteID string, from, to time.Time) ([]model.Sample, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,site_id,collector,condition,status,notes,collected_at,created_at,updated_at FROM samples WHERE site_id=? AND collected_at>=? AND collected_at<? ORDER BY collected_at`, siteID, encodeTime(from), encodeTime(to))
	if err != nil {
		return nil, fmt.Errorf("samples collected between: %w", err)
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

func (s *Store) EventTimeline(ctx context.Context, sampleID string) ([]model.Event, error) {
	return s.ListEvents(ctx, sampleID, 500)
}
