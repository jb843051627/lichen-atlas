package store

import (
	"context"
	"fmt"
)

func (s *Store) initSchema(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS sites (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, region TEXT NOT NULL,
			elevation_m REAL NOT NULL, status TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS samples (
			id TEXT PRIMARY KEY, site_id TEXT NOT NULL, collector TEXT NOT NULL, condition TEXT NOT NULL,
			status TEXT NOT NULL, notes TEXT NOT NULL, collected_at TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
			FOREIGN KEY(site_id) REFERENCES sites(id)
		)`,
		`CREATE TABLE IF NOT EXISTS readings (
			id TEXT PRIMARY KEY, sample_id TEXT NOT NULL, kind TEXT NOT NULL, value REAL NOT NULL,
			unit TEXT NOT NULL, recorded_at TEXT NOT NULL, FOREIGN KEY(sample_id) REFERENCES samples(id)
		)`,
		`CREATE TABLE IF NOT EXISTS taxa (
			id TEXT PRIMARY KEY, scientific TEXT NOT NULL, common_name TEXT NOT NULL,
			rank TEXT NOT NULL, authority TEXT NOT NULL, default_status TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS identifications (
			id TEXT PRIMARY KEY, sample_id TEXT NOT NULL, taxon_id TEXT NOT NULL, reviewer TEXT NOT NULL,
			confidence REAL NOT NULL, status TEXT NOT NULL, comment TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
			FOREIGN KEY(sample_id) REFERENCES samples(id), FOREIGN KEY(taxon_id) REFERENCES taxa(id)
		)`,
		`CREATE TABLE IF NOT EXISTS reviews (
			id TEXT PRIMARY KEY, sample_id TEXT NOT NULL, reviewer TEXT NOT NULL, decision TEXT NOT NULL,
			reason TEXT NOT NULL, created_at TEXT NOT NULL, FOREIGN KEY(sample_id) REFERENCES samples(id)
		)`,
		`CREATE TABLE IF NOT EXISTS archives (
			id TEXT PRIMARY KEY, sample_id TEXT NOT NULL UNIQUE, box_code TEXT NOT NULL, sealed_by TEXT NOT NULL,
			seal_state TEXT NOT NULL, sealed_at TEXT NOT NULL, note TEXT NOT NULL, FOREIGN KEY(sample_id) REFERENCES samples(id)
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY, sample_id TEXT NOT NULL, kind TEXT NOT NULL, state TEXT NOT NULL,
			attempts INTEGER NOT NULL, available_at TEXT NOT NULL, claimed_at TEXT NOT NULL, finished_at TEXT NOT NULL,
			last_error TEXT NOT NULL, FOREIGN KEY(sample_id) REFERENCES samples(id)
		)`,
		`CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY, sample_id TEXT NOT NULL, event_type TEXT NOT NULL,
			payload TEXT NOT NULL, created_at TEXT NOT NULL, FOREIGN KEY(sample_id) REFERENCES samples(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_samples_site ON samples(site_id, collected_at)`,
		`CREATE INDEX IF NOT EXISTS idx_readings_sample ON readings(sample_id, recorded_at)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_ready ON tasks(state, available_at)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("initialize schema: %w", err)
		}
	}
	return nil
}
