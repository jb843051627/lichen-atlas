package model

import (
	"fmt"
	"strings"
)

type Taxon struct {
	ID            string
	Scientific    string
	CommonName    string
	Rank          string
	Authority     string
	DefaultStatus string
}

func (t Taxon) Validate() error {
	if strings.TrimSpace(t.ID) == "" || strings.TrimSpace(t.Scientific) == "" {
		return fmt.Errorf("taxon identity is required")
	}
	if strings.TrimSpace(t.Rank) == "" {
		return fmt.Errorf("taxon rank is required")
	}
	return nil
}

func (t Taxon) IsSpecies() bool { return strings.EqualFold(t.Rank, "species") }
