package store

import (
	"context"
	"fmt"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Store) AddReview(ctx context.Context, value model.Review) error {
	if err := value.Validate(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO reviews(id,sample_id,reviewer,decision,reason,created_at) VALUES(?,?,?,?,?,?)`,
		value.ID, value.SampleID, value.Reviewer, value.Decision, value.Reason, encodeTime(value.CreatedAt))
	if err != nil {
		return fmt.Errorf("add review: %w", err)
	}
	return nil
}

func (s *Store) ListReviews(ctx context.Context, sampleID string) ([]model.Review, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,sample_id,reviewer,decision,reason,created_at FROM reviews WHERE sample_id=? ORDER BY created_at`, sampleID)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()
	result := make([]model.Review, 0)
	for rows.Next() {
		var value model.Review
		var created string
		if err := rows.Scan(&value.ID, &value.SampleID, &value.Reviewer, &value.Decision, &value.Reason, &created); err != nil {
			return nil, err
		}
		value.CreatedAt = decodeTime(created)
		result = append(result, value)
	}
	return result, rows.Err()
}
