package store

import (
	"context"
	"fmt"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Store) AppendEvent(ctx context.Context, event model.Event) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO events(id,sample_id,event_type,payload,created_at) VALUES(?,?,?,?,?)`,
		event.ID, event.SampleID, event.EventType, event.Payload, encodeTime(event.CreatedAt))
	if err != nil {
		return fmt.Errorf("append event: %w", err)
	}
	return nil
}

func (s *Store) ListEvents(ctx context.Context, sampleID string, limit int) ([]model.Event, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,sample_id,event_type,payload,created_at FROM events WHERE sample_id=? ORDER BY created_at DESC LIMIT ?`, sampleID, limit)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	result := make([]model.Event, 0)
	for rows.Next() {
		var event model.Event
		var created string
		if err := rows.Scan(&event.ID, &event.SampleID, &event.EventType, &event.Payload, &created); err != nil {
			return nil, err
		}
		event.CreatedAt = decodeTime(created)
		result = append(result, event)
	}
	return result, rows.Err()
}

func (s *Store) CountEvents(ctx context.Context, sampleID string) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events WHERE sample_id=?`, sampleID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
