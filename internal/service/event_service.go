package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type EventService struct {
	append func(context.Context, model.Event) error
}

func NewEventService(appendEvent func(context.Context, model.Event) error) *EventService {
	return &EventService{append: appendEvent}
}

func (s *EventService) Emit(ctx context.Context, event model.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if event.ID == "" || event.SampleID == "" || strings.TrimSpace(event.EventType) == "" {
		return fmt.Errorf("event identity is required")
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if s.append == nil {
		return fmt.Errorf("event sink is unavailable")
	}
	return s.append(ctx, event)
}

func (s *EventService) EmitState(ctx context.Context, sampleID, state string, at time.Time) error {
	return s.Emit(ctx, model.Event{ID: "state-" + sampleID + "-" + state, SampleID: sampleID, EventType: "sample." + state, CreatedAt: at, Payload: "{}"})
}

func SortEvents(events []model.Event) []model.Event {
	result := append([]model.Event(nil), events...)
	sort.SliceStable(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result
}

func EventsOfType(events []model.Event, eventType string) []model.Event {
	result := make([]model.Event, 0)
	for _, event := range events {
		if event.EventType == eventType {
			result = append(result, event)
		}
	}
	return SortEvents(result)
}

func LatestEvent(events []model.Event) (model.Event, bool) {
	ordered := SortEvents(events)
	if len(ordered) == 0 {
		return model.Event{}, false
	}
	return ordered[len(ordered)-1], true
}

func EventCountByType(events []model.Event) map[string]int {
	result := make(map[string]int)
	for _, event := range events {
		result[event.EventType]++
	}
	return result
}
