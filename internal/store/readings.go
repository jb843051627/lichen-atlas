package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type readingCache struct {
	mu    sync.RWMutex
	items map[string][]model.Reading
}

var cacheByStore sync.Map

func (s *Store) getReadingCache() *readingCache {
	value, _ := cacheByStore.LoadOrStore(s, &readingCache{items: make(map[string][]model.Reading)})
	return value.(*readingCache)
}

func (s *Store) AddReading(ctx context.Context, reading model.Reading) error {
	if err := reading.Validate(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO readings(id,sample_id,kind,value,unit,recorded_at) VALUES(?,?,?,?,?,?)`,
		reading.ID, reading.SampleID, reading.Kind, reading.Value, reading.Unit, encodeTime(reading.RecordedAt))
	if err != nil {
		return fmt.Errorf("add reading: %w", err)
	}
	c := s.getReadingCache()
	c.mu.Lock()
	c.items[reading.SampleID] = append(c.items[reading.SampleID], reading)
	c.mu.Unlock()
	return nil
}

func cloneReadings(source []model.Reading) []model.Reading {
	result := make([]model.Reading, len(source))
	copy(result, source)
	return result
}

func (s *Store) ListReadings(ctx context.Context, sampleID string) ([]model.Reading, error) {
	c := s.getReadingCache()
	c.mu.RLock()
	if cached, ok := c.items[sampleID]; ok {
		result := cloneReadings(cached)
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()
	rows, err := s.db.QueryContext(ctx, `SELECT id,sample_id,kind,value,unit,recorded_at FROM readings WHERE sample_id=? ORDER BY recorded_at`, sampleID)
	if err != nil {
		return nil, fmt.Errorf("list readings: %w", err)
	}
	defer rows.Close()
	result := make([]model.Reading, 0)
	for rows.Next() {
		var reading model.Reading
		var recorded string
		if err := rows.Scan(&reading.ID, &reading.SampleID, &reading.Kind, &reading.Value, &reading.Unit, &recorded); err != nil {
			return nil, err
		}
		reading.RecordedAt = decodeTime(recorded)
		result = append(result, reading)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.items[sampleID] = cloneReadings(result)
	c.mu.Unlock()
	return cloneReadings(result), nil
}

func (s *Store) CountReadings(ctx context.Context, sampleID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM readings WHERE sample_id=?`, sampleID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count readings: %w", err)
	}
	return count, nil
}

func (s *Store) ClearReadingCache(sampleID string) {
	c := s.getReadingCache()
	c.mu.Lock()
	delete(c.items, sampleID)
	c.mu.Unlock()
}
