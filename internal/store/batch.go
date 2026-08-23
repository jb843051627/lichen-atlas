package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type SampleBundle struct {
	Sample         model.Sample
	Readings       []model.Reading
	Identification *model.Identification
	Reviews        []model.Review
	Archive        *model.ArchiveRecord
}

func (s *Store) LoadSampleBundle(ctx context.Context, sampleID string) (SampleBundle, error) {
	sample, err := s.GetSample(ctx, sampleID)
	if err != nil {
		return SampleBundle{}, fmt.Errorf("load bundle sample: %w", err)
	}
	readings, err := s.ListReadings(ctx, sampleID)
	if err != nil {
		return SampleBundle{}, err
	}
	identification, identErr := s.GetIdentification(ctx, sampleID)
	if identErr != nil && identErr != ErrNotFound {
		return SampleBundle{}, identErr
	}
	reviews, err := s.ListReviews(ctx, sampleID)
	if err != nil {
		return SampleBundle{}, err
	}
	archive, archiveErr := s.GetArchive(ctx, sampleID)
	if archiveErr != nil && archiveErr != ErrNotFound {
		return SampleBundle{}, archiveErr
	}
	return SampleBundle{Sample: *sample, Readings: readings, Identification: identification, Reviews: reviews, Archive: archive}, nil
}

func (s *Store) LoadSiteBundles(ctx context.Context, siteID string, limit int) ([]SampleBundle, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM samples WHERE site_id=? ORDER BY collected_at LIMIT ?`, siteID, limit)
	if err != nil {
		return nil, fmt.Errorf("list site bundle ids: %w", err)
	}
	defer rows.Close()
	result := make([]SampleBundle, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		bundle, err := s.LoadSampleBundle(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, bundle)
	}
	return result, rows.Err()
}

func (s *Store) InsertSampleWithLocation(ctx context.Context, sample model.Sample, payload string) error {
	if err := sample.Validate(); err != nil {
		return err
	}
	return s.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO samples(id,site_id,collector,condition,status,notes,collected_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, sample.ID, sample.SiteID, sample.Collector, sample.Condition, sample.Status, sample.Notes, encodeTime(sample.CollectedAt), encodeTime(sample.CreatedAt), encodeTime(sample.UpdatedAt)); err != nil {
			return fmt.Errorf("insert sample: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO events(id,sample_id,event_type,payload,created_at) VALUES(?,?,?,?,?)`, "event-location-"+sample.ID, sample.ID, "sample.location", payload, encodeTime(sample.CreatedAt)); err != nil {
			return fmt.Errorf("insert sample location: %w", err)
		}
		return nil
	})
}

func (s *Store) CreateSampleWithEvent(ctx context.Context, sample model.Sample, event model.Event) error {
	if err := sample.Validate(); err != nil {
		return err
	}
	return s.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO samples(id,site_id,collector,condition,status,notes,collected_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, sample.ID, sample.SiteID, sample.Collector, sample.Condition, sample.Status, sample.Notes, encodeTime(sample.CollectedAt), encodeTime(sample.CreatedAt), encodeTime(sample.UpdatedAt)); err != nil {
			return fmt.Errorf("insert sample: %w", err)
		}
		return AppendEventTx(tx, model.Event{ID: event.ID, SampleID: event.SampleID, EventType: event.EventType, Payload: "", CreatedAt: event.CreatedAt})
	})
}

func (s *Store) MoveSampleStateWithEvent(ctx context.Context, id, from, to, eventID, payload string, at time.Time) error {
	return s.WithTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE samples SET status=?,updated_at=? WHERE id=? AND status=?`, to, encodeTime(at), id, from)
		if err != nil {
			return fmt.Errorf("move sample state: %w", err)
		}
		n, _ := result.RowsAffected()
		if n != 1 {
			return ErrConflict
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO events(id,sample_id,event_type,payload,created_at) VALUES(?,?,?,?,?)`, eventID, id, "sample."+to, payload, encodeTime(at)); err != nil {
			return fmt.Errorf("append state event: %w", err)
		}
		return nil
	})
}

func (s *Store) DeleteSample(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrNotFound
	}
	return s.WithTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM samples WHERE id=?`, id)
		if err != nil {
			return err
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			return ErrNotFound
		}
		return nil
	})
}
