package service

import (
	"context"
	"time"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
)

func runWorker(ctx context.Context, process func(context.Context) (bool, error), failureMessage string) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for ctx.Err() == nil {
		processed, err := process(ctx)
		if err != nil && ctx.Err() == nil {
			sharedlog.Error(failureMessage, "err", err)
		}
		if processed {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
	return nil
}
