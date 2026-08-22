package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type Observation struct {
	ID          string
	SampleID    string
	PointCode   string
	Observer    string
	Surface     string
	CoveragePct float64
	Color       string
	Texture     string
	NotedAt     time.Time
}

func (o Observation) Validate() error {
	if o.ID == "" || o.SampleID == "" || o.PointCode == "" || strings.TrimSpace(o.Observer) == "" {
		return fmt.Errorf("observation identity is required")
	}
	if o.CoveragePct < 0 || o.CoveragePct > 100 {
		return fmt.Errorf("coverage is outside range")
	}
	if o.NotedAt.IsZero() {
		return fmt.Errorf("observation time is required")
	}
	return nil
}

type ObservationBatch struct {
	Items     []Observation
	StartedAt time.Time
	EndedAt   time.Time
}

func (b ObservationBatch) Validate() error {
	if len(b.Items) == 0 {
		return fmt.Errorf("observation batch is empty")
	}
	seen := make(map[string]bool)
	for _, item := range b.Items {
		if err := item.Validate(); err != nil {
			return err
		}
		if seen[item.ID] {
			return fmt.Errorf("observation %s is repeated", item.ID)
		}
		seen[item.ID] = true
	}
	return nil
}

func GroupObservations(items []Observation) map[string][]Observation {
	result := make(map[string][]Observation)
	for _, item := range items {
		result[item.SampleID] = append(result[item.SampleID], item)
	}
	for key := range result {
		result[key] = SortObservations(result[key])
	}
	return result
}

func SortObservations(items []Observation) []Observation {
	result := append([]Observation(nil), items...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].NotedAt.Equal(result[j].NotedAt) {
			return result[i].ID < result[j].ID
		}
		return result[i].NotedAt.Before(result[j].NotedAt)
	})
	return result
}

func MeanCoverage(items []Observation) float64 {
	if len(items) == 0 {
		return 0
	}
	total := 0.0
	for _, item := range items {
		total += item.CoveragePct
	}
	return total / float64(len(items))
}

func DominantSurface(items []Observation) string {
	counts := make(map[string]int)
	for _, item := range items {
		counts[strings.ToLower(item.Surface)]++
	}
	best, bestCount := "", 0
	for surface, count := range counts {
		if count > bestCount || (count == bestCount && surface < best) {
			best, bestCount = surface, count
		}
	}
	return best
}

func ObserveWithContext(ctx context.Context, points []TransectPoint, observe func(context.Context, TransectPoint) (Observation, error)) ([]Observation, error) {
	result := make([]Observation, 0, len(points))
	for _, point := range SortTransectPoints(points) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		item, err := observe(ctx, point)
		if err != nil {
			return nil, fmt.Errorf("observe %s: %w", point.Code, err)
		}
		result = append(result, item)
	}
	return result, nil
}

func MeasurementWindow(items []model.Reading, start, end time.Time) []model.Reading {
	result := make([]model.Reading, 0)
	for _, item := range items {
		if (start.IsZero() || !item.RecordedAt.Before(start)) && (end.IsZero() || item.RecordedAt.Before(end)) {
			result = append(result, item)
		}
	}
	return SortReadingsByTime(result)
}

func ObservationAge(now time.Time, item Observation) time.Duration {
	if item.NotedAt.IsZero() || now.Before(item.NotedAt) {
		return 0
	}
	return now.Sub(item.NotedAt)
}
