package model

import (
	"fmt"
	"strings"
	"time"
)

type Review struct {
	ID        string
	SampleID  string
	Reviewer  string
	Decision  string
	Reason    string
	CreatedAt time.Time
}

func (r Review) Validate() error {
	if r.ID == "" || r.SampleID == "" || strings.TrimSpace(r.Reviewer) == "" {
		return fmt.Errorf("review identity is required")
	}
	if r.Decision != "approve" && r.Decision != "reject" {
		return fmt.Errorf("review decision is invalid")
	}
	if r.Decision == "reject" && strings.TrimSpace(r.Reason) == "" {
		return fmt.Errorf("rejection reason is required")
	}
	return nil
}

func (r Review) IsApproval() bool { return r.Decision == "approve" }
