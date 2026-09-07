package runtime

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"

	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
)

func StartKafkaConsumer(ctx context.Context, wg *sync.WaitGroup, run func(ctx context.Context, brokers []string) error) {
	raw := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	if raw == "" {
		sharedlog.Info("KAFKA_BOOTSTRAP_SERVERS not set; skipping Kafka consumer")
		return
	}

	brokers := messaging.ParseBootstrapServers(raw)
	if len(brokers) == 0 {
		sharedlog.Info("KAFKA_BOOTSTRAP_SERVERS has no brokers after parsing; skipping Kafka consumer")
		return
	}

	wg.Go(func() {
		for ctx.Err() == nil {
			err := run(ctx, brokers)
			if ctx.Err() != nil || errors.Is(err, context.Canceled) {
				return
			}
			sharedlog.Error("kafka consumer stopped; restarting in 5s", "err", err)
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	})
}
