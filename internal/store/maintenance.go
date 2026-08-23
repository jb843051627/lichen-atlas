package store

import (
	"context"
	"fmt"
	"time"
)

type MaintenanceRecord struct {
	ID           string
	InstrumentID string
	Kind         string
	PerformedBy  string
	PerformedAt  time.Time
	NextDue      time.Time
	Note         string
}

func (s *Store) SaveMaintenance(ctx context.Context, record MaintenanceRecord) error {
	if record.ID == "" || record.InstrumentID == "" || record.Kind == "" {
		return fmt.Errorf("maintenance identity is required")
	}
	if record.NextDue.Before(record.PerformedAt) {
		return fmt.Errorf("next due date precedes maintenance")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO events(id,sample_id,event_type,payload,created_at) VALUES(?,?,?,?,?)`, record.ID, record.InstrumentID, "instrument.maintenance", record.Note, encodeTime(record.PerformedAt))
	if err != nil {
		return fmt.Errorf("save maintenance: %w", err)
	}
	return nil
}

func (s *Store) ListMaintenance(ctx context.Context, instrumentID string) ([]MaintenanceRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,sample_id,event_type,payload,created_at FROM events WHERE sample_id=? AND event_type='instrument.maintenance' ORDER BY created_at DESC`, instrumentID)
	if err != nil {
		return nil, fmt.Errorf("list maintenance: %w", err)
	}
	defer rows.Close()
	result := make([]MaintenanceRecord, 0)
	for rows.Next() {
		var record MaintenanceRecord
		var created string
		if err := rows.Scan(&record.ID, &record.InstrumentID, &record.Kind, &record.Note, &created); err != nil {
			return nil, err
		}
		record.PerformedAt = decodeTime(created)
		result = append(result, record)
	}
	return result, rows.Err()
}
