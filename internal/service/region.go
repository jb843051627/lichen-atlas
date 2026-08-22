package service

import "strings"

func NormalizeRegionName(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func RegionMatches(value, query string) bool {
	return query == "" || NormalizeRegionName(value) == NormalizeRegionName(query)
}

func RegionKnown(value string) bool { return strings.TrimSpace(value) != "" }
