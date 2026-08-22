package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type TransectPoint struct {
	Code      string
	Label     string
	DistanceM float64
	Latitude  float64
	Longitude float64
	VisitedAt time.Time
	SampleIDs []string
}

type Transect struct {
	ID         string
	SiteID     string
	Name       string
	Points     []TransectPoint
	StartedAt  time.Time
	FinishedAt time.Time
}

func (t Transect) Validate() error {
	if strings.TrimSpace(t.ID) == "" || strings.TrimSpace(t.SiteID) == "" {
		return fmt.Errorf("transect identity is required")
	}
	if len(t.Points) == 0 {
		return fmt.Errorf("transect needs at least one point")
	}
	seen := make(map[string]bool)
	for _, point := range t.Points {
		if point.Code == "" || seen[point.Code] {
			return fmt.Errorf("transect point code is empty or repeated")
		}
		seen[point.Code] = true
		if point.DistanceM < 0 {
			return fmt.Errorf("transect distance cannot be negative")
		}
	}
	return nil
}

func SortTransectPoints(points []TransectPoint) []TransectPoint {
	result := append([]TransectPoint(nil), points...)
	sort.SliceStable(result, func(i, j int) bool { return result[i].DistanceM < result[j].DistanceM })
	return result
}

func TransectDistance(points []TransectPoint) float64 {
	if len(points) < 2 {
		return 0
	}
	sorted := SortTransectPoints(points)
	return sorted[len(sorted)-1].DistanceM - sorted[0].DistanceM
}

func VisitedPoints(points []TransectPoint) int {
	count := 0
	for _, point := range points {
		if !point.VisitedAt.IsZero() {
			count++
		}
	}
	return count
}

func TransectProgress(points []TransectPoint) float64 {
	if len(points) == 0 {
		return 0
	}
	return float64(VisitedPoints(points)) / float64(len(points))
}

func CollectTransect(ctx context.Context, transect Transect, collect func(context.Context, TransectPoint) ([]model.Reading, error)) (map[string][]model.Reading, error) {
	if err := transect.Validate(); err != nil {
		return nil, err
	}
	points := SortTransectPoints(transect.Points)
	result := make(map[string][]model.Reading, len(points))
	for _, point := range points {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("collect transect %s: %w", transect.ID, err)
		}
		readings, err := collect(ctx, point)
		if err != nil {
			return nil, fmt.Errorf("collect point %s: %w", point.Code, err)
		}
		result[point.Code] = append([]model.Reading(nil), readings...)
	}
	return result, nil
}

type TransectBuffer struct {
	mu     sync.RWMutex
	points map[string]TransectPoint
}

func NewTransectBuffer() *TransectBuffer {
	return &TransectBuffer{points: make(map[string]TransectPoint)}
}

func (b *TransectBuffer) Put(point TransectPoint) {
	b.mu.Lock()
	b.points[point.Code] = point
	b.mu.Unlock()
}

func (b *TransectBuffer) Get(code string) (TransectPoint, bool) {
	b.mu.RLock()
	point, ok := b.points[code]
	b.mu.RUnlock()
	return point, ok
}

func (b *TransectBuffer) Snapshot() []TransectPoint {
	result := make([]TransectPoint, 0, len(b.points))
	for _, point := range b.points {
		point.SampleIDs = append([]string(nil), point.SampleIDs...)
		result = append(result, point)
	}
	return SortTransectPoints(result)
}

func (b *TransectBuffer) Reset() {
	b.mu.Lock()
	b.points = make(map[string]TransectPoint)
	b.mu.Unlock()
}
