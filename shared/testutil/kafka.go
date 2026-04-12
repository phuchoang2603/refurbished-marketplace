package testutil

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

type KafkaContainer struct {
	*kafka.KafkaContainer
}

func SetupKafka(t *testing.T) *KafkaContainer {
	t.Helper()

	ctx := context.Background()
	c, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0")
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
