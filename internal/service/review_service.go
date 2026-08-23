package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type ReviewSummary struct {
	Total      int
	Approvals  int
	Rejections int
	Latest     model.Review
	HasLatest  bool
}

func SummarizeReviews(reviews []model.Review) ReviewSummary {
	ordered := append([]model.Review(nil), reviews...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].CreatedAt.Before(ordered[j].CreatedAt) })
	result := ReviewSummary{Total: len(ordered)}
	if len(ordered) > 0 {
		result.Latest, result.HasLatest = ordered[len(ordered)-1], true
	}
	for _, review := range ordered {
		if review.IsApproval() {
			result.Approvals++
		} else if review.Decision == "reject" {
			result.Rejections++
		}
	}
	return result
}

func ReviewDecision(reviews []model.Review) string {
	summary := SummarizeReviews(reviews)
	if !summary.HasLatest {
		return "pending"
	}
	return summary.Latest.Decision
}

func ValidateReviewChain(reviews []model.Review) error {
	if len(reviews) == 0 {
		return fmt.Errorf("review chain is empty")
	}
	seen := make(map[string]bool)
	for _, review := range reviews {
		if err := review.Validate(); err != nil {
			return err
		}
		if seen[review.ID] {
			return fmt.Errorf("review %s is repeated", review.ID)
		}
		seen[review.ID] = true
	}
	return nil
}

func ReviewCanAdvance(reviews []model.Review, requiredConfidence float64) bool {
	for _, review := range reviews {
		if review.IsApproval() && requiredConfidence <= 1 {
			return true
		}
	}
	return false
}

func ReviewReasonText(review model.Review) string {
	if review.IsApproval() {
		return "approved"
	}
	if strings.TrimSpace(review.Reason) == "" {
		return "rejected without a reason"
	}
	return strings.TrimSpace(review.Reason)
}
