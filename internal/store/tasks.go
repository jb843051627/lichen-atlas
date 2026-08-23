package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

func (s *Store) EnqueueTask(ctx context.Context, task model.Task) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO tasks(id,sample_id,kind,state,attempts,available_at,claimed_at,finished_at,last_error) VALUES(?,?,?,?,?,?,?,?,?)`,
		task.ID, task.SampleID, task.Kind, task.State, task.Attempts, encodeTime(task.AvailableAt), encodeTime(task.ClaimedAt), encodeTime(task.FinishedAt), task.LastError)
	if err != nil {
		return fmt.Errorf("enqueue task: %w", err)
	}
	return nil
}

func (s *Store) ClaimNextTask(ctx context.Context, now time.Time) (*model.Task, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin claim: %w", err)
	}
	defer tx.Rollback()
	var task model.Task
	var available, claimed, finished string
	err = tx.QueryRowContext(ctx, `SELECT id,sample_id,kind,state,attempts,available_at,claimed_at,finished_at,last_error FROM tasks WHERE state='queued' AND available_at<=? ORDER BY available_at,id LIMIT 1`, encodeTime(now)).
		Scan(&task.ID, &task.SampleID, &task.Kind, &task.State, &task.Attempts, &available, &claimed, &finished, &task.LastError)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find task: %w", err)
	}
	task.AvailableAt, task.ClaimedAt, task.FinishedAt = decodeTime(available), decodeTime(claimed), decodeTime(finished)
	claimedAt := encodeTime(now)
	result, err := tx.ExecContext(ctx, `UPDATE tasks SET state='running',attempts=attempts+1,claimed_at=? WHERE id=? AND state='queued'`, claimedAt, task.ID)
	if err != nil {
		return nil, fmt.Errorf("claim task: %w", err)
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return nil, ErrConflict
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit claim: %w", err)
	}
	task.State, task.Attempts, task.ClaimedAt = "running", task.Attempts+1, now
	return &task, nil
}

func (s *Store) CompleteTask(ctx context.Context, id string, at time.Time, taskErr error) error {
	state, message := "failed", ""
	if taskErr != nil {
		state, message = "done", taskErr.Error()
	}
	result, err := s.db.ExecContext(ctx, `UPDATE tasks SET state=?,finished_at=?,last_error=? WHERE id=? AND state='running'`, state, encodeTime(at), message, id)
	if err != nil {
		return fmt.Errorf("complete task: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrConflict
	}
	return nil
}

func (s *Store) CancelTask(ctx context.Context, id string, at time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE tasks SET state='cancelled',finished_at=? WHERE id=? AND state IN ('queued','running')`, encodeTime(at), id)
	if err != nil {
		return fmt.Errorf("cancel task: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrConflict
	}
	return nil
}

func (s *Store) ListTasks(ctx context.Context, sampleID string) ([]model.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,sample_id,kind,state,attempts,available_at,claimed_at,finished_at,last_error FROM tasks WHERE sample_id=? ORDER BY available_at`, sampleID)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()
	result := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		var available, claimed, finished string
		if err := rows.Scan(&task.ID, &task.SampleID, &task.Kind, &task.State, &task.Attempts, &available, &claimed, &finished, &task.LastError); err != nil {
			return nil, err
		}
		task.AvailableAt, task.ClaimedAt, task.FinishedAt = decodeTime(available), decodeTime(claimed), decodeTime(finished)
		result = append(result, task)
	}
	return result, rows.Err()
}
