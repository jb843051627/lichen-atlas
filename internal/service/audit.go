package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type AuditEntry struct {
	ID        string
	SampleID  string
	Actor     string
	Action    string
	Detail    string
	CreatedAt time.Time
}

func (e AuditEntry) Validate() error {
	if e.ID == "" || e.SampleID == "" || strings.TrimSpace(e.Actor) == "" || strings.TrimSpace(e.Action) == "" {
		return fmt.Errorf("audit entry identity is required")
	}
	if e.CreatedAt.IsZero() {
		return fmt.Errorf("audit entry time is required")
	}
	return nil
}

func BuildAuditEntry(sample model.Sample, actor, action, detail string, at time.Time) AuditEntry {
	return AuditEntry{ID: "audit-" + sample.ID + "-" + action, SampleID: sample.ID, Actor: actor, Action: action, Detail: detail, CreatedAt: at}
}

func SortAudit(entries []AuditEntry) []AuditEntry {
	result := append([]AuditEntry(nil), entries...)
	sort.SliceStable(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result
}

func AuditActions(entries []AuditEntry) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, entry := range entries {
		if !seen[entry.Action] {
			seen[entry.Action] = true
			result = append(result, entry.Action)
		}
	}
	sort.Strings(result)
	return result
}

func AuditHas(entries []AuditEntry, action string) bool {
	for _, entry := range entries {
		if entry.Action == action {
			return true
		}
	}
	return false
}
