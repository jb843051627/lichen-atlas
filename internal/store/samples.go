package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Store) CreateSample(ctx context.Context, sample model.Sample) error {
	if err := sample.Validate(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO samples(id,site_id,collector,condition,status,notes,collected_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`,
		sample.ID, sample.SiteID, sample.Collector, sample.Condition, sample.Status, sample.Notes,
		encodeTime(sample.CollectedAt), encodeTime(sample.CreatedAt), encodeTime(sample.UpdatedAt))
	if err != nil {
		return fmt.Errorf("create sample: %w", err)
	}
	return nil
}

func (s *Store) GetSample(ctx context.Context, id string) (*model.Sample, error) {
	var sample model.Sample
	var collected, created, updated string
	err := s.db.QueryRowContext(ctx, `SELECT id,site_id,collector,condition,status,notes,collected_at,created_at,updated_at FROM samples WHERE id=?`, id).
		Scan(&sample.ID, &sample.SiteID, &sample.Collector, &sample.Condition, &sample.Status, &sample.Notes, &collected, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get sample: %w", err)
	}
	sample.CollectedAt, sample.CreatedAt, sample.UpdatedAt = decodeTime(collected), decodeTime(created), decodeTime(updated)
	return &sample, nil
}

func (s *Store) ListSamplesBySite(ctx context.Context, siteID string) ([]model.Sample, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,site_id,collector,condition,status,notes,collected_at,created_at,updated_at FROM samples WHERE site_id=? ORDER BY collected_at`, siteID)
	if err != nil {
		return nil, fmt.Errorf("list samples: %w", err)
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

func (s *Store) UpdateSampleState(ctx context.Context, id, from, to, at string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE samples SET status=?,updated_at=? WHERE id=? AND status=?`, to, at, id, from)
	if err != nil {
		return fmt.Errorf("update sample state: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		_, lookupErr := s.GetSample(ctx, id)
		if lookupErr != nil {
			return lookupErr
		}
		return ErrConflict
	}
	return nil
}

func (s *Store) UpdateSampleNotes(ctx context.Context, id, notes, at string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE samples SET notes=?,updated_at=? WHERE id=?`, notes, at, id)
	if err != nil {
		return fmt.Errorf("update sample notes: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
