package store

import (
	"context"
	"fmt"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Store) AddFieldNote(ctx context.Context, note model.FieldNote) error {
	if note.IsEmpty() {
		return fmt.Errorf("field note is empty")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO events(id,sample_id,event_type,payload,created_at) VALUES(?,?,?,?,?)`, note.ID, note.SampleID, "field.note", note.Text, encodeTime(note.CreatedAt))
	if err != nil {
		return fmt.Errorf("add field note: %w", err)
	}
	return nil
}

func (s *Store) CountNotes(ctx context.Context, sampleID string) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events WHERE sample_id=? AND event_type='field.note'`, sampleID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
