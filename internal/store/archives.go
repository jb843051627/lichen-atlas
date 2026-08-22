package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Store) CreateArchive(ctx context.Context, value model.ArchiveRecord) error {
	if err := value.Validate(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO archives(id,sample_id,box_code,sealed_by,seal_state,sealed_at,note) VALUES(?,?,?,?,?,?,?)`,
		value.ID, value.SampleID, value.BoxCode, value.SealedBy, value.SealState, encodeTime(value.SealedAt), value.Note)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	return nil
}

func (s *Store) GetArchive(ctx context.Context, sampleID string) (*model.ArchiveRecord, error) {
	var value model.ArchiveRecord
	var sealed string
	err := s.db.QueryRowContext(ctx, `SELECT id,sample_id,box_code,sealed_by,seal_state,sealed_at,note FROM archives WHERE sample_id=?`, sampleID).
		Scan(&value.ID, &value.SampleID, &value.BoxCode, &value.SealedBy, &value.SealState, &sealed, &value.Note)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get archive: %w", err)
	}
	value.SealedAt = decodeTime(sealed)
	return &value, nil
}

func (s *Store) UpdateArchiveState(ctx context.Context, sampleID, state, note string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE archives SET seal_state=?,note=? WHERE sample_id=?`, state, note, sampleID)
	if err != nil {
		return fmt.Errorf("update archive: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
