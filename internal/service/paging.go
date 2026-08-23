package service

import (
	"fmt"
	"strings"
)

type Page struct {
	Offset int
	Limit  int
}

func (p Page) Validate() error {
	if p.Offset < 0 {
		return fmt.Errorf("page offset cannot be negative")
	}
	if p.Limit < 0 || p.Limit > 500 {
		return fmt.Errorf("page limit must be between zero and 500")
	}
	return nil
}

func ParsePage(offset, limit int) (Page, error) {
	page := Page{Offset: offset, Limit: limit}
	if page.Limit == 0 {
		page.Limit = 100
	}
	return page, page.Validate()
}

func NormalizeSearch(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

func ContainsAny(value string, terms []string) bool {
	value = NormalizeSearch(value)
	for _, term := range terms {
		if term != "" && strings.Contains(value, NormalizeSearch(term)) {
			return true
		}
	}
	return false
}
