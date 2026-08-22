package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type VisitService struct {
	mu     sync.RWMutex
	visits map[string]model.FieldVisit
	marks  map[string]model.SampleMark
}

func NewVisitService() *VisitService {
	return &VisitService{visits: make(map[string]model.FieldVisit), marks: make(map[string]model.SampleMark)}
}

func (s *VisitService) Open(ctx context.Context, visit model.FieldVisit) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := visit.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.visits[visit.ID]; exists {
		return fmt.Errorf("field visit %s already exists", visit.ID)
	}
	s.visits[visit.ID] = visit
	return nil
}

func (s *VisitService) Close(ctx context.Context, id string, at time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	visit, ok := s.visits[id]
	if !ok {
		return fmt.Errorf("field visit %s not found", id)
	}
	if !visit.EndedAt.IsZero() {
		return fmt.Errorf("field visit %s is already closed", id)
	}
	if at.Before(visit.StartedAt) {
		return fmt.Errorf("field visit closes before it starts")
	}
	visit.EndedAt = at
	s.visits[id] = visit
	return nil
}

func (s *VisitService) Get(ctx context.Context, id string) (model.FieldVisit, error) {
	if err := ctx.Err(); err != nil {
		return model.FieldVisit{}, err
	}
	s.mu.RLock()
	visit, ok := s.visits[id]
	s.mu.RUnlock()
	if !ok {
		return model.FieldVisit{}, fmt.Errorf("field visit %s not found", id)
	}
	return visit, nil
}

func (s *VisitService) Mark(ctx context.Context, mark model.SampleMark) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := mark.Validate(); err != nil {
		return err
	}
	if _, err := s.Get(ctx, mark.VisitID); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.marks[mark.Code]; exists {
		return fmt.Errorf("sample mark %s already exists", mark.Code)
	}
	s.marks[mark.Code] = mark
	return nil
}

func (s *VisitService) FindMarks(ctx context.Context, visitID string) []model.SampleMark {
	if ctx.Err() != nil {
		return nil
	}
	s.mu.RLock()
	result := make([]model.SampleMark, 0)
	for _, mark := range s.marks {
		if visitID == "" || mark.VisitID == visitID {
			result = append(result, mark)
		}
	}
	s.mu.RUnlock()
	return result
}

func (s *VisitService) RenameTeam(ctx context.Context, id, team string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	team = strings.TrimSpace(team)
	if team == "" {
		return fmt.Errorf("team is empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	visit, ok := s.visits[id]
	if !ok {
		return fmt.Errorf("field visit %s not found", id)
	}
	visit.Team = team
	s.visits[id] = visit
	return nil
}
