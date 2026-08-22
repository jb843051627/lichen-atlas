package service

import "strings"

func HumanStatus(value string) string {
	switch strings.ToLower(value) {
	case "draft":
		return "field draft"
	case "measured":
		return "measurements complete"
	case "identified":
		return "identification ready"
	case "archived":
		return "sealed archive"
	default:
		return "unknown"
	}
}

func IsTerminalStatus(value string) bool {
	return value == "archived" || value == "retired" || value == "cancelled"
}

func IsKnownSampleStatus(value string) bool {
	switch value {
	case "draft", "measured", "identified", "archived":
		return true
	default:
		return false
	}
}

func StatusOrder(value string) int {
	switch value {
	case "draft":
		return 1
	case "measured":
		return 2
	case "identified":
		return 3
	case "archived":
		return 4
	default:
		return 0
	}
}
