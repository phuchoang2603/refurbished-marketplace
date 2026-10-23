package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestWorkerDrainsThenPolls(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
	}{
		{name: "idle"},
		{name: "error", err: errors.New("store unavailable")},
	} {
		t.Run(test.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				var calls atomic.Int32
				done := make(chan error, 1)
				go func() {
					done <- runWorker(ctx, func(context.Context) (bool, error) {
						return calls.Add(1) <= 3, test.err
					}, "test worker failed")
				}()
				synctest.Wait()
				if got := calls.Load(); got != 4 {
					t.Fatalf("expected backlog to drain immediately, got %d calls", got)
				}
				time.Sleep(time.Second)
				synctest.Wait()
				if got := calls.Load(); got != 5 {
					t.Fatalf("expected one idle poll, got %d calls", got)
				}
				cancel()
				if err := <-done; err != nil {
					t.Fatal(err)
				}
			})
		})
	}
}

func TestWorkerStopsWhenBusy(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	calls := 0
	err := runWorker(ctx, func(context.Context) (bool, error) {
		calls++
		cancel()
		return true, nil
	}, "test worker failed")
	if err != nil || calls != 1 {
		t.Fatalf("busy worker did not stop on cancellation: calls=%d, err=%v", calls, err)
	}
}

func TestWorkerSkipsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := runWorker(ctx, func(context.Context) (bool, error) {
		t.Error("worker processed work after cancellation")
		return false, nil
	}, "test worker failed"); err != nil {
		t.Fatal(err)
	}
}
