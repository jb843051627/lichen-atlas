package service

import "strings"

type ProtocolRule struct {
	Code        string
	Description string
	Scope       string
	Required    bool
}

var fieldProtocol = []ProtocolRule{
	{Code: "site_open", Description: "sampling site must be open", Scope: "site", Required: true},
	{Code: "collector_named", Description: "collector must be traceable", Scope: "sample", Required: true},
	{Code: "timestamp_present", Description: "reading needs a field timestamp", Scope: "reading", Required: true},
	{Code: "environment_complete", Description: "three environmental families are required", Scope: "reading", Required: true},
	{Code: "duplicate_reading", Description: "same-kind readings at one instant need review", Scope: "reading", Required: false},
	{Code: "confidence_bound", Description: "confidence stays between zero and one", Scope: "identification", Required: true},
	{Code: "species_high_confidence", Description: "high confidence requires species rank", Scope: "identification", Required: true},
	{Code: "reviewer_named", Description: "reviewer is recorded separately", Scope: "review", Required: true},
	{Code: "approval_before_seal", Description: "an approved review precedes sealing", Scope: "archive", Required: true},
	{Code: "box_code_unique", Description: "archive box code cannot repeat", Scope: "archive", Required: true},
	{Code: "archive_event", Description: "sealing emits a state event", Scope: "archive", Required: true},
	{Code: "task_retry_limit", Description: "background tasks have bounded retries", Scope: "task", Required: true},
	{Code: "cancel_propagates", Description: "cancellation stops a field batch", Scope: "context", Required: true},
	{Code: "report_snapshot", Description: "report reads a consistent sample set", Scope: "report", Required: true},
	{Code: "timezone_explicit", Description: "report time keeps survey location context", Scope: "report", Required: true},
	{Code: "raw_payload_kept", Description: "raw field payload is not overwritten", Scope: "event", Required: true},
	{Code: "stationary_sensor", Description: "fixed sensors identify their source", Scope: "reading", Required: false},
	{Code: "weather_gap", Description: "long measurement gaps are marked", Scope: "reading", Required: false},
	{Code: "calibration_age", Description: "old instrument readings need review", Scope: "reading", Required: false},
	{Code: "sample_reopen", Description: "rejected identification can return to pending", Scope: "state", Required: true},
	{Code: "retired_site", Description: "retired sites only expose historical data", Scope: "site", Required: true},
	{Code: "drying_state", Description: "specimen drying state is tracked", Scope: "archive", Required: false},
	{Code: "field_note", Description: "field notes remain sample events", Scope: "event", Required: true},
	{Code: "location_accuracy", Description: "coordinate accuracy travels with the sample", Scope: "sample", Required: true},
	{Code: "operator_separation", Description: "review and sealing operators are distinguishable", Scope: "archive", Required: true},
	{Code: "report_limit", Description: "report exports have a sample limit", Scope: "report", Required: true},
	{Code: "worker_shutdown", Description: "service shutdown releases background workers", Scope: "worker", Required: true},
	{Code: "worker_recover", Description: "one failed task cannot kill the worker", Scope: "worker", Required: true},
	{Code: "cache_copy", Description: "service does not expose cache backing slices", Scope: "cache", Required: true},
	{Code: "error_chain", Description: "boundary errors retain sentinel meaning", Scope: "error", Required: true},
	{Code: "transaction_atomic", Description: "archive state and record commit atomically", Scope: "transaction", Required: true},
	{Code: "db_restart", Description: "survey data survives process restart", Scope: "persistence", Required: true},
}

func ProtocolForScope(scope string) []ProtocolRule {
	result := make([]ProtocolRule, 0)
	for _, rule := range fieldProtocol {
		if scope == "" || strings.EqualFold(rule.Scope, scope) {
			result = append(result, rule)
		}
	}
	return result
}

func ProtocolHas(code string) bool {
	for _, rule := range fieldProtocol {
		if strings.EqualFold(rule.Code, code) {
			return true
		}
	}
	return false
}

func RequiredProtocolRules() []ProtocolRule {
	result := make([]ProtocolRule, 0)
	for _, rule := range fieldProtocol {
		if rule.Required {
			result = append(result, rule)
		}
	}
	return result
}

func RequiredScopes() []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, rule := range fieldProtocol {
		if !seen[rule.Scope] {
			seen[rule.Scope] = true
			result = append(result, rule.Scope)
		}
	}
	return result
}

func ProtocolCoverage(codes []string) float64 {
	if len(fieldProtocol) == 0 {
		return 1
	}
	seen := make(map[string]bool)
	for _, code := range codes {
		if ProtocolHas(code) {
			seen[strings.ToLower(code)] = true
		}
	}
	return float64(len(seen)) / float64(len(fieldProtocol))
}

func MissingProtocol(codes []string) []ProtocolRule {
	seen := make(map[string]bool)
	for _, code := range codes {
		seen[strings.ToLower(code)] = true
	}
	result := make([]ProtocolRule, 0)
	for _, rule := range fieldProtocol {
		if rule.Required && !seen[strings.ToLower(rule.Code)] {
			result = append(result, rule)
		}
	}
	return result
}

func ScopeForProtocol(code string) string {
	for _, rule := range fieldProtocol {
		if strings.EqualFold(rule.Code, code) {
			return rule.Scope
		}
	}
	return "unknown"
}
