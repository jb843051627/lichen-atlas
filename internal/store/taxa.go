package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Store) CreateTaxon(ctx context.Context, taxon model.Taxon) error {
	if err := taxon.Validate(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO taxa(id,scientific,common_name,rank,authority,default_status) VALUES(?,?,?,?,?,?)`,
		taxon.ID, taxon.Scientific, taxon.CommonName, taxon.Rank, taxon.Authority, taxon.DefaultStatus)
	if err != nil {
		return fmt.Errorf("create taxon: %w", err)
	}
	return nil
}

func (s *Store) GetTaxon(ctx context.Context, id string) (*model.Taxon, error) {
	var taxon model.Taxon
	err := s.db.QueryRowContext(ctx, `SELECT id,scientific,common_name,rank,authority,default_status FROM taxa WHERE id=?`, id).
		Scan(&taxon.ID, &taxon.Scientific, &taxon.CommonName, &taxon.Rank, &taxon.Authority, &taxon.DefaultStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get taxon: %w", err)
	}
	return &taxon, nil
}

func (s *Store) ListTaxa(ctx context.Context, rank string) ([]model.Taxon, error) {
	query, args := `SELECT id,scientific,common_name,rank,authority,default_status FROM taxa ORDER BY scientific`, []any{}
	if rank != "" {
		query = `SELECT id,scientific,common_name,rank,authority,default_status FROM taxa WHERE rank=? ORDER BY scientific`
		args = []any{rank}
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list taxa: %w", err)
	}
	defer rows.Close()
	result := make([]model.Taxon, 0)
	for rows.Next() {
		var taxon model.Taxon
		if err := rows.Scan(&taxon.ID, &taxon.Scientific, &taxon.CommonName, &taxon.Rank, &taxon.Authority, &taxon.DefaultStatus); err != nil {
			return nil, err
		}
		result = append(result, taxon)
	}
	return result, rows.Err()
}
