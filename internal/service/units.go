package service

import "strings"

func UnitCompatible(kind, unit string) bool {
	rule, ok := FindEnvironmentRule(kind)
	return ok && strings.EqualFold(rule.Unit, unit)
}

func FamilyKinds(family string) []string {
	result := make([]string, 0)
	for _, rule := range environmentRules {
		if rule.Family == family {
			result = append(result, rule.Kind)
		}
	}
	return result
}

func IsEnvironmentalKind(kind string) bool {
	_, ok := FindEnvironmentRule(kind)
	return ok
}
