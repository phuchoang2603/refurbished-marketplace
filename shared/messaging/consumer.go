package messaging

import (
	"context"
	"fmt"

	sharedtrace "refurbished-marketplace/shared/trace"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

type KafkaMessage struct {
	Topic     string
	Partition int32
	Offset    int64
	Value     []byte
	Headers   map[string]string
}

func KafkaMessageID(m KafkaMessage) string {
	return fmt.Sprintf("%s/%d/%d", m.Topic, m.Partition, m.Offset)
}

type KafkaHandler func(ctx context.Context, msg KafkaMessage) error

type KafkaConsumerConfig struct {
	BootstrapServers []string
	GroupID          string
	Topics           []string
	// TracerName scopes consumer process spans; empty uses "kafka-consumer".
	TracerName string
}

type KafkaConsumer struct {
	client     *kgo.Client
	handler    KafkaHandler
	tracerName string
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
	name := cfg.TracerName
	if name == "" {
		name = "kafka-consumer"
	}
	return &KafkaConsumer{client: cl, handler: h, tracerName: name}, nil
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
					headers := headersFromRecord(r)
					msgCtx := sharedtrace.ContextFromHeaders(gctx, headers)
					msgCtx, span := sharedtrace.Tracer(c.tracerName).Start(
						msgCtx, "messaging process "+r.Topic,
						trace.WithSpanKind(trace.SpanKindConsumer),
						trace.WithAttributes(
							attribute.String("messaging.system", "kafka"),
							attribute.String("messaging.destination.name", r.Topic),
							attribute.String("messaging.operation.type", "process"),
							attribute.Int("messaging.kafka.offset", int(r.Offset)),
							attribute.Int("messaging.kafka.destination.partition", int(r.Partition)),
						),
					)
					err := c.handler(msgCtx, KafkaMessage{
						Topic: r.Topic, Partition: r.Partition, Offset: r.Offset, Value: r.Value, Headers: headers,
					})
					if err != nil {
						span.RecordError(err)
						span.SetStatus(codes.Error, err.Error())
					}
					span.End()
					if err != nil {
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

func headersFromRecord(r *kgo.Record) map[string]string {
	if r == nil || len(r.Headers) == 0 {
		return nil
	}
	out := make(map[string]string, len(r.Headers))
	for _, h := range r.Headers {
		out[h.Key] = string(h.Value)
	}
	return out
}
