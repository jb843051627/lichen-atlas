package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type ReplayState struct {
	SampleID      string
	Status        string
	ReadingCount  int
	HasIdentified bool
	HasReview     bool
	HasArchive    bool
	LastEventID   string
}

func ReplayEvents(events []model.Event) (ReplayState, error) {
	if len(events) == 0 {
		return ReplayState{}, fmt.Errorf("event stream is empty")
	}
	ordered := append([]model.Event(nil), events...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].CreatedAt.Before(ordered[j].CreatedAt) })
	state := ReplayState{SampleID: ordered[0].SampleID, Status: model.SampleDraft}
	for _, event := range ordered {
		if event.SampleID != state.SampleID {
			return ReplayState{}, fmt.Errorf("event stream contains multiple samples")
		}
		if err := applyEvent(&state, event); err != nil {
			return ReplayState{}, err
		}
		state.LastEventID = event.ID
	}
	return state, nil
}

func applyEvent(state *ReplayState, event model.Event) error {
	switch event.EventType {
	case "sample.created":
		if state.Status != model.SampleDraft {
			return fmt.Errorf("sample created event is out of order")
		}
	case "sample.measured":
		state.Status = model.SampleMeasured
	case "sample.identified":
		state.Status, state.HasIdentified = model.SampleIdentified, true
	case "sample.archived":
		state.Status, state.HasArchive = model.SampleArchived, true
	case "reading.added":
		state.ReadingCount++
	case "identification.reviewed":
		state.HasReview = true
	case "field.note":
		if strings.TrimSpace(event.Payload) == "" {
			return fmt.Errorf("field note event has empty payload")
		}
	default:
		return fmt.Errorf("unknown sample event %q", event.EventType)
	}
	return nil
}

func ReplayConsistent(state ReplayState, sample model.Sample) bool {
	return state.SampleID == sample.ID && state.Status == sample.Status
}
