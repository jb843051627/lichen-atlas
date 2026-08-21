package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Store) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func InsertArchiveTx(tx *sql.Tx, value model.ArchiveRecord) error {
	_, err := tx.Exec(`INSERT INTO archives(id,sample_id,box_code,sealed_by,seal_state,sealed_at,note) VALUES(?,?,?,?,?,?,?)`,
		value.ID, value.SampleID, value.BoxCode, value.SealedBy, value.SealState, encodeTime(value.SealedAt), value.Note)
	if err != nil {
		return fmt.Errorf("insert archive in transaction: %w", err)
	}
	return nil
}

func UpdateSampleStateTx(tx *sql.Tx, id, from, to, at string) error {
	result, err := tx.Exec(`UPDATE samples SET status=?,updated_at=? WHERE id=? AND status=?`, to, at, id, from)
	if err != nil {
		return fmt.Errorf("update sample in transaction: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrConflict
	}
	return nil
}
