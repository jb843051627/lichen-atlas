package service

import "context"

func WaitForQueue(ctx context.Context, ready <-chan struct{}) error {
	select {
	case <-ready:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func DrainQueue[T any](ctx context.Context, input <-chan T, handle func(T) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case value, ok := <-input:
			if !ok {
				return nil
			}
			if err := handle(value); err != nil {
				return err
			}
		}
	}
}
