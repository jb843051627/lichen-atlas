package service

import (
	"fmt"
	"strings"
)

func ValidateCollector(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("collector is empty")
	}
	if len([]rune(name)) > 80 {
		return fmt.Errorf("collector name is too long")
	}
	return nil
}

func NormalizeRegion(region string) string { return strings.TrimSpace(strings.ToLower(region)) }
