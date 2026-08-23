package model

import (
	"fmt"
	"strings"
	"time"
)

type ArchiveRecord struct {
	ID        string
	SampleID  string
	BoxCode   string
	SealedBy  string
	SealState string
	SealedAt  time.Time
	Note      string
}

func (a ArchiveRecord) Validate() error {
	if a.ID == "" || a.SampleID == "" || strings.TrimSpace(a.BoxCode) == "" {
		return fmt.Errorf("archive identity is required")
	}
	if strings.TrimSpace(a.SealedBy) == "" {
		return fmt.Errorf("archive operator is required")
	}
	return nil
}

func (a ArchiveRecord) IsSealed() bool { return a.SealState == "sealed" }
