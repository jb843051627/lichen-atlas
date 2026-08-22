package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type LifecycleAction struct {
	From      string
	To        string
	EventType string
	Required  []string
}

var lifecycleActions = map[string]LifecycleAction{
	"measure":  {From: model.SampleDraft, To: model.SampleMeasured, EventType: "sample.measured", Required: []string{"reading"}},
	"identify": {From: model.SampleMeasured, To: model.SampleIdentified, EventType: "sample.identified", Required: []string{"reading", "taxon"}},
	"archive":  {From: model.SampleIdentified, To: model.SampleArchived, EventType: "sample.archived", Required: []string{"accepted_review", "quality"}},
}

func LifecycleActionFor(name string) (LifecycleAction, bool) {
	action, ok := lifecycleActions[name]
	return action, ok
}

func ValidateLifecycleAction(ctx context.Context, sample model.Sample, actionName string, report model.SampleReport) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	action, ok := LifecycleActionFor(actionName)
	if !ok {
		return fmt.Errorf("unknown lifecycle action %s", actionName)
	}
	if sample.Status != action.From {
		return fmt.Errorf("sample state %s does not allow %s", sample.Status, actionName)
	}
	if actionName == "measure" && len(report.Readings) < 3 {
		return fmt.Errorf("measurement action needs three readings")
	}
	if actionName == "identify" && report.Identification == nil {
		return fmt.Errorf("identification action needs an identification")
	}
	if actionName == "archive" {
		score := SampleQuality(sample, report.Readings, report.Identification, report.Reviews)
		if !CanSeal(score, report.Identification, report.Reviews) {
			return fmt.Errorf("archive quality score %.2f is insufficient", score)
		}
	}
	return nil
}

func NextLifecycleAction(sample model.Sample) string {
	switch sample.Status {
	case model.SampleDraft:
		return "measure"
	case model.SampleMeasured:
		return "identify"
	case model.SampleIdentified:
		return "archive"
	default:
		return ""
	}
}

func LifecycleDeadline(start time.Time, action string) time.Time {
	durations := map[string]time.Duration{"measure": 8 * time.Hour, "identify": 72 * time.Hour, "archive": 14 * 24 * time.Hour}
	return start.Add(durations[action])
}

func IsLifecycleLate(now, started time.Time, action string) bool {
	deadline := LifecycleDeadline(started, action)
	return !deadline.IsZero() && now.After(deadline)
}
