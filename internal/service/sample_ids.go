package service

import "strings"

func SampleIDValid(value string) bool { return strings.TrimSpace(value) != "" }

func SampleIDPrefix(value string) string {
	if len(value) < 3 {
		return value
	}
	return value[:3]
}
