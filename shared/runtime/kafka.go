package runtime

import (
	"context"
	"errors"
	"os"
	"sync"

	sharedlog "refurbished-marketplace/shared/observe/log"

	"refurbished-marketplace/shared/messaging"
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
		if err := run(ctx, brokers); err != nil && !errors.Is(err, context.Canceled) {
			sharedlog.Error("kafka consumer stopped", "err", err)
		}
	})
}
