package store

import (
	"context"
	"fmt"
)

type SampleCounts struct {
	Total      int
	Draft      int
	Measured   int
	Identified int
	Archived   int
}

func (s *Store) CountSamples(ctx context.Context, siteID string) (SampleCounts, error) {
	var result SampleCounts
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*),
		SUM(CASE WHEN status='draft' THEN 1 ELSE 0 END),
		SUM(CASE WHEN status='measured' THEN 1 ELSE 0 END),
		SUM(CASE WHEN status='identified' THEN 1 ELSE 0 END),
		SUM(CASE WHEN status='archived' THEN 1 ELSE 0 END)
		FROM samples WHERE site_id=?`, siteID).Scan(&result.Total, &result.Draft, &result.Measured, &result.Identified, &result.Archived)
	if err != nil {
		return SampleCounts{}, fmt.Errorf("count samples: %w", err)
	}
	return result, nil
}

func (s *Store) Ping(ctx context.Context) error {
	var value int
	return s.db.QueryRowContext(ctx, `SELECT 1`).Scan(&value)
}
