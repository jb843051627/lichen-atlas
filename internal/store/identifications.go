package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Store) SaveIdentification(ctx context.Context, value model.Identification) error {
	if err := value.Validate(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO identifications(id,sample_id,taxon_id,reviewer,confidence,status,comment,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`,
		value.ID, value.SampleID, value.TaxonID, value.Reviewer, value.Confidence, value.Status, value.Comment,
		encodeTime(value.CreatedAt), encodeTime(value.UpdatedAt))
	if err != nil {
		return fmt.Errorf("save identification: %w", err)
	}
	return nil
}

func (s *Store) GetIdentification(ctx context.Context, sampleID string) (*model.Identification, error) {
	var value model.Identification
	var created, updated string
	err := s.db.QueryRowContext(ctx, `SELECT id,sample_id,taxon_id,reviewer,confidence,status,comment,created_at,updated_at FROM identifications WHERE sample_id=? ORDER BY created_at DESC LIMIT 1`, sampleID).
		Scan(&value.ID, &value.SampleID, &value.TaxonID, &value.Reviewer, &value.Confidence, &value.Status, &value.Comment, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get identification: %w", err)
	}
	value.CreatedAt, value.UpdatedAt = decodeTime(created), decodeTime(updated)
	return &value, nil
}

func (s *Store) UpdateIdentificationStatus(ctx context.Context, id, status, comment, at string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE identifications SET status=?,comment=?,updated_at=? WHERE id=?`, status, comment, at, id)
	if err != nil {
		return fmt.Errorf("update identification: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
