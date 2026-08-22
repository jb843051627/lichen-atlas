package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Store) CreateSite(ctx context.Context, site model.Site) error {
	if err := site.Validate(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO sites(id,name,region,elevation_m,status,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`,
		site.ID, site.Name, site.Region, site.ElevationM, site.Status, encodeTime(site.CreatedAt), encodeTime(site.UpdatedAt))
	if err != nil {
		return fmt.Errorf("create site: %w", err)
	}
	return nil
}

func (s *Store) GetSite(ctx context.Context, id string) (*model.Site, error) {
	var site model.Site
	var created, updated string
	err := s.db.QueryRowContext(ctx, `SELECT id,name,region,elevation_m,status,created_at,updated_at FROM sites WHERE id=?`, id).
		Scan(&site.ID, &site.Name, &site.Region, &site.ElevationM, &site.Status, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get site: %w", err)
	}
	site.CreatedAt, site.UpdatedAt = decodeTime(created), decodeTime(updated)
	return &site, nil
}

func (s *Store) ListSites(ctx context.Context, region string) ([]model.Site, error) {
	query, args := `SELECT id,name,region,elevation_m,status,created_at,updated_at FROM sites ORDER BY name`, []any{}
	if region != "" {
		query = `SELECT id,name,region,elevation_m,status,created_at,updated_at FROM sites WHERE region=? ORDER BY name`
		args = []any{region}
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list sites: %w", err)
	}
	defer rows.Close()
	result := make([]model.Site, 0)
	for rows.Next() {
		var site model.Site
		var created, updated string
		if err := rows.Scan(&site.ID, &site.Name, &site.Region, &site.ElevationM, &site.Status, &created, &updated); err != nil {
			return nil, err
		}
		site.CreatedAt, site.UpdatedAt = decodeTime(created), decodeTime(updated)
		result = append(result, site)
	}
	return result, rows.Err()
}

func (s *Store) UpdateSiteStatus(ctx context.Context, id, status string, at string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE sites SET status=?,updated_at=? WHERE id=?`, status, at, id)
	if err != nil {
		return fmt.Errorf("update site status: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
