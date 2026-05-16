package testutil

import (
	"context"
	"errors"
	"testing"
	"time"

	"refurbished-marketplace/shared/messaging"

	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaContainer struct {
	*kafka.KafkaContainer
}

func SetupKafka(t *testing.T) *KafkaContainer {
	t.Helper()

	ctx := context.Background()
	c, err := kafka.Run(ctx, "confluentinc/confluent-local:8.2.0")
	if err != nil {
		t.Fatalf("start kafka container: %v", err)
	}

	t.Cleanup(func() {
		if err := c.Terminate(ctx); err != nil {
			t.Fatalf("terminate kafka container: %v", err)
		}
	})

	return &KafkaContainer{KafkaContainer: c}
}

func ProduceKafkaRecord(t *testing.T, ctx context.Context, brokers []string, topic string, value []byte) {
	t.Helper()

	prod, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer prod.Close()

	res := prod.ProduceSync(ctx, &kgo.Record{Topic: topic, Value: value})
	if err := res.FirstErr(); err != nil {
		t.Fatalf("ProduceSync: %v", err)
	}
}

func StartKafkaConsumer(t *testing.T, ctx context.Context, brokers []string, groupID string, topics []string, handler messaging.KafkaHandler) (context.CancelFunc, <-chan error) {
	t.Helper()

	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: brokers,
		GroupID:          groupID,
		Topics:           topics,
	}, handler)
	if err != nil {
		t.Fatalf("NewKafkaConsumer: %v", err)
	}
	t.Cleanup(func() { _ = consumer.Close() })

	runCtx, cancel := context.WithCancel(ctx)
	errCh := make(chan error, 1)
	go func() {
		errCh <- consumer.Run(runCtx)
	}()

	return cancel, errCh
}

func WaitForKafkaCondition(
	t *testing.T,
	errCh <-chan error,
	cancel context.CancelFunc,
	timeout, interval time.Duration,
	timeoutMsg string,
	condition func() (bool, error),
) {
	t.Helper()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	deadline := time.After(timeout)

	for {
		select {
		case err := <-errCh:
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Fatalf("Consumer exited unexpectedly: %v", err)
			}
			return
		case <-deadline:
			t.Fatal(timeoutMsg)
		case <-ticker.C:
			done, err := condition()
			if err != nil {
				t.Fatal(err)
			}
			if done {
				cancel()
				return
			}
		}
	}
}
