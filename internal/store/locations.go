package store

import (
	"context"
	"fmt"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Store) SaveLocation(ctx context.Context, sampleID string, location model.Location) error {
	if err := location.Validate(); err != nil {
		return err
	}
	payload := fmt.Sprintf("{\"lat\":%.6f,\"lon\":%.6f,\"accuracy_m\":%.2f}", location.Latitude, location.Longitude, location.AccuracyM)
	_, err := s.db.ExecContext(ctx, `INSERT INTO events(id,sample_id,event_type,payload,created_at) VALUES(?,?,?,?,datetime('now'))`, "location-"+sampleID, sampleID, "sample.location", payload)
	return err
}

func (s *Store) HasLocation(ctx context.Context, sampleID string) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events WHERE sample_id=? AND event_type='sample.location'`, sampleID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}
