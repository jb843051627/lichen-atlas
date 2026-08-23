package store

import (
	"context"
	"fmt"
	"time"
)

func (s *Store) PruneEvents(ctx context.Context, before time.Time) (int64, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE created_at<? AND event_type NOT IN ('sample.created','sample.archived')`, encodeTime(before))
	if err != nil {
		return 0, fmt.Errorf("prune events: %w", err)
	}
	return result.RowsAffected()
}

func (s *Store) Vacuum(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `VACUUM`)
	return err
}

func (s *Store) Checkpoint(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `PRAGMA wal_checkpoint(PASSIVE)`)
	return err
}

func (s *Store) TableCounts(ctx context.Context) (map[string]int, error) {
	result := make(map[string]int)
	for _, table := range []string{"sites", "samples", "readings", "taxa", "identifications", "reviews", "archives", "tasks", "events"} {
		var count int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+table).Scan(&count); err != nil {
			return nil, fmt.Errorf("count table %s: %w", table, err)
		}
		result[table] = count
	}
	return result, nil
}
