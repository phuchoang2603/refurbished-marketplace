package messaging

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaMessage struct {
	Topic     string
	Partition int32
	Offset    int64
	Value     []byte
}

func KafkaMessageID(m KafkaMessage) string {
	return fmt.Sprintf("%s/%d/%d", m.Topic, m.Partition, m.Offset)
}

type KafkaHandler func(ctx context.Context, msg KafkaMessage) error

type KafkaConsumerConfig struct {
	BootstrapServers string
	GroupID          string
	Topics           []string
	PreferIPv4       bool
}

// KafkaConsumer wraps a single-poller consumer with manual commits.
type KafkaConsumer struct {
	c       *kafka.Consumer
	handler KafkaHandler
}

// NewKafkaConsumer builds a consumer subscribed to Topics. Poll interval is fixed at 250ms.
func NewKafkaConsumer(cfg KafkaConsumerConfig, h KafkaHandler) (*KafkaConsumer, error) {
	if cfg.BootstrapServers == "" {
		return nil, fmt.Errorf("bootstrapServers is required")
	}
	if cfg.GroupID == "" {
		return nil, fmt.Errorf("groupID is required")
	}
	if len(cfg.Topics) == 0 {
		return nil, fmt.Errorf("topics is required")
	}
	if h == nil {
		return nil, fmt.Errorf("handler is required")
	}

	cm := kafka.ConfigMap{
		"bootstrap.servers":        cfg.BootstrapServers,
		"group.id":                 cfg.GroupID,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false,
		"enable.auto.offset.store": false,
	}
	if cfg.PreferIPv4 {
		_ = cm.SetKey("broker.address.family", "v4")
		_ = cm.SetKey("socket.connection.setup.timeout.ms", "60000")
	}

	c, err := kafka.NewConsumer(&cm)
	if err != nil {
		return nil, err
	}
	if err := c.SubscribeTopics(cfg.Topics, nil); err != nil {
		_ = c.Close()
		return nil, err
	}
	return &KafkaConsumer{c: c, handler: h}, nil
}

func (c *KafkaConsumer) Close() error {
	if c.c == nil {
		return nil
	}
	return c.c.Close()
}

// Run polls until ctx is cancelled. One goroutine should call Poll (do not fan out Poll).
func (c *KafkaConsumer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ev := c.c.Poll(250)
		switch e := ev.(type) {
		case nil:
			continue
		case *kafka.Message:
			topic := ""
			if e.TopicPartition.Topic != nil {
				topic = *e.TopicPartition.Topic
			}
			msg := KafkaMessage{
				Topic:     topic,
				Partition: e.TopicPartition.Partition,
				Offset:    int64(e.TopicPartition.Offset),
				Value:     e.Value,
			}
			if err := c.handler(ctx, msg); err != nil {
				continue
			}
			if _, err := c.c.CommitMessage(e); err != nil {
				continue
			}
		case kafka.Error:
			return e
		}
	}
}
