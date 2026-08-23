package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type FieldAlert struct {
	Code      string
	Severity  string
	SampleID  string
	Message   string
	CreatedAt time.Time
	Resolved  bool
}

func NewReadingAlert(reading model.Reading, message string, at time.Time) FieldAlert {
	severity := "notice"
	if rule, ok := FindEnvironmentRule(reading.Kind); ok {
		if reading.Value < rule.Min || reading.Value > rule.Max {
			severity = "critical"
		}
	}
	return FieldAlert{Code: "reading-" + reading.ID, Severity: severity, SampleID: reading.SampleID, Message: message, CreatedAt: at}
}

func ValidateAlert(alert FieldAlert) error {
	if alert.Code == "" || alert.SampleID == "" || strings.TrimSpace(alert.Message) == "" {
		return fmt.Errorf("field alert identity is required")
	}
	if alert.Severity != "notice" && alert.Severity != "warning" && alert.Severity != "critical" {
		return fmt.Errorf("field alert severity is invalid")
	}
	if alert.CreatedAt.IsZero() {
		return fmt.Errorf("field alert time is required")
	}
	return nil
}

func SortAlerts(alerts []FieldAlert) []FieldAlert {
	result := append([]FieldAlert(nil), alerts...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Severity == result[j].Severity {
			return result[i].CreatedAt.Before(result[j].CreatedAt)
		}
		return alertRank(result[i].Severity) > alertRank(result[j].Severity)
	})
	return result
}

func alertRank(severity string) int {
	switch severity {
	case "critical":
		return 3
	case "warning":
		return 2
	case "notice":
		return 1
	default:
		return 0
	}
}

func UnresolvedAlerts(alerts []FieldAlert) []FieldAlert {
	result := make([]FieldAlert, 0)
	for _, alert := range alerts {
		if !alert.Resolved {
			result = append(result, alert)
		}
	}
	return SortAlerts(result)
}

func ResolveAlert(alert FieldAlert) (FieldAlert, error) {
	if err := ValidateAlert(alert); err != nil {
		return FieldAlert{}, err
	}
	alert.Resolved = true
	return alert, nil
}
