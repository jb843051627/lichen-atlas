package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jb843051627/lichen-atlas/internal/model"
)

type TaskRunner struct {
	Input   <-chan model.Task
	Output  chan<- model.Task
	Workers int
	Timeout time.Duration
}

func (r TaskRunner) Run(ctx context.Context, handle func(context.Context, model.Task) error) error {
	if r.Workers <= 0 {
		r.Workers = 1
	}
	if r.Timeout <= 0 {
		r.Timeout = 30 * time.Second
	}
	if r.Input == nil {
		return fmt.Errorf("task input is nil")
	}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errors := make(chan error, r.Workers)
	var wg sync.WaitGroup
	for i := 0; i < r.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-workerCtx.Done():
					return
				case task, ok := <-r.Input:
					if !ok {
						return
					}
					taskCtx, taskCancel := context.WithTimeout(workerCtx, r.Timeout)
					err := handle(taskCtx, task)
					taskCancel()
					if err != nil {
						select {
						case errors <- err:
						default:
						}
						cancel()
						return
					}
					if r.Output != nil {
						select {
						case r.Output <- task:
						case <-workerCtx.Done():
							return
						}
					}
				}
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
		select {
		case err := <-errors:
			return err
		default:
			return nil
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

func RetryDelay(attempt int, base time.Duration) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if base <= 0 {
		base = time.Second
	}
	if attempt > 6 {
		attempt = 6
	}
	return base * time.Duration(1<<(attempt-1))
}

func ShouldRetry(task model.Task, err error) bool {
	return err != nil && task.Attempts < 3 && !task.IsTerminal()
}
