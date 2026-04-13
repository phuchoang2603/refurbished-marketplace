package messaging

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"golang.org/x/sync/errgroup"
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
	BootstrapServers []string
	GroupID          string
	Topics           []string
}

type KafkaConsumer struct {
	client  *kgo.Client
	handler KafkaHandler
}

func NewKafkaConsumer(cfg KafkaConsumerConfig, h KafkaHandler) (*KafkaConsumer, error) {
	if len(cfg.BootstrapServers) == 0 || cfg.GroupID == "" || len(cfg.Topics) == 0 || h == nil {
		return nil, fmt.Errorf("bootstrapServers, groupID, topics and handler are required")
	}

	cl, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.BootstrapServers...),
		kgo.ConsumerGroup(cfg.GroupID),
		kgo.ConsumeTopics(cfg.Topics...),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, err
	}
	return &KafkaConsumer{client: cl, handler: h}, nil
}

func (c *KafkaConsumer) Close() error {
	if c.client != nil {
		c.client.Close()
	}
	return nil
}

func (c *KafkaConsumer) Run(ctx context.Context) error {
	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return nil
		}
		if err := fetches.Err(); err != nil {
			return err
		}
		if fetches.Empty() {
			continue
		}

		g, gctx := errgroup.WithContext(ctx)
		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			g.Go(func() error {
				for _, r := range p.Records {
					if err := c.handler(gctx, KafkaMessage{
						Topic: r.Topic, Partition: r.Partition, Offset: r.Offset, Value: r.Value,
					}); err != nil {
						return err
					}
				}
				return nil
			})
		})
		if err := g.Wait(); err != nil {
			continue
		}
		if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
			continue
		}
	}
}
