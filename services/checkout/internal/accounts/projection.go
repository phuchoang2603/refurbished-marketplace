package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	projectionKey = "checkout:accounts:v1"
	offsetField   = "_next_offset"
)

var ErrCheckpointChanged = errors.New("account projection checkpoint changed; reload from Redis")

type redisStore interface {
	HMGet(context.Context, string, ...string) *redis.SliceCmd
	HGet(context.Context, string, string) *redis.StringCmd
	Eval(context.Context, string, []string, ...interface{}) *redis.Cmd
}

type Projection struct {
	store redisStore
}

func New(client *redis.Client) *Projection {
	return &Projection{store: client}
}

func (p *Projection) CreatedAt(ctx context.Context, ids ...string) ([]time.Time, error) {
	values, err := p.store.HMGet(ctx, projectionKey, ids...).Result()
	if err != nil {
		return nil, fmt.Errorf("read account projection: %w", err)
	}
	times := make([]time.Time, len(ids))
	for i, value := range values {
		text, ok := value.(string)
		if !ok {
			return nil, errors.New("account not projected yet")
		}
		created, err := time.Parse(time.RFC3339Nano, text)
		if err != nil || created.IsZero() {
			return nil, errors.New("invalid projected account time")
		}
		times[i] = created.UTC()
	}
	return times, nil
}

func (p *Projection) nextOffset(ctx context.Context) (int64, error) {
	value, err := p.store.HGet(ctx, projectionKey, offsetField).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	offset, err := strconv.ParseInt(value, 10, 64)
	if err != nil || offset < 0 {
		return 0, errors.New("invalid account projection checkpoint")
	}
	return offset, nil
}

// Facts and their checkpoint share one non-expiring, non-evictable hash so Redis
// recovery cannot retain an offset while losing the corresponding account facts.
const applyBatch = `
if (redis.call('HGET', KEYS[1], '_next_offset') or '0') ~= ARGV[1] then
    return 0
end
redis.call('HSET', KEYS[1], unpack(ARGV, 2))
return 1
`

func (p *Projection) apply(ctx context.Context, expected int64, records []*kgo.Record) (int64, error) {
	if len(records) == 0 {
		return expected, nil
	}
	args := []interface{}{strconv.FormatInt(expected, 10)}
	next := expected
	for _, record := range records {
		if record.Topic != messaging.EventTypeAccountCreated || record.Partition != 0 || record.Offset < next {
			return expected, errors.New("invalid account event position")
		}
		var event messaging.AccountCreated
		if err := json.Unmarshal(record.Value, &event); err != nil {
			return expected, fmt.Errorf("decode account event: %w", err)
		}
		id, err := uuid.Parse(event.UserID)
		if err != nil || id == uuid.Nil || id.String() != event.UserID || event.CreatedAt.IsZero() {
			return expected, errors.New("invalid account facts")
		}
		args = append(args, event.UserID, event.CreatedAt.UTC().Format(time.RFC3339Nano))
		next = record.Offset + 1
	}
	args = append(args, offsetField, strconv.FormatInt(next, 10))
	applied, err := p.store.Eval(ctx, applyBatch, []string{projectionKey}, args...).Int()
	if err != nil {
		return expected, fmt.Errorf("persist account projection: %w", err)
	}
	if applied != 1 {
		return expected, ErrCheckpointChanged
	}
	return next, nil
}

func (p *Projection) Run(ctx context.Context, brokers []string) error {
	next, err := p.nextOffset(ctx)
	if err != nil {
		return err
	}
	// The retained account topic has one partition. Redis owns the checkpoint;
	// Kafka group commits would skip facts after Redis data loss or rollback.
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumePartitions(map[string]map[int32]kgo.Offset{
			messaging.EventTypeAccountCreated: {0: kgo.NewOffset().At(next)},
		}),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		return err
	}
	defer client.Close()
	for ctx.Err() == nil {
		stored, err := p.nextOffset(ctx)
		if err != nil {
			return err
		}
		if stored != next {
			return ErrCheckpointChanged
		}
		pollCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		fetches := client.PollRecords(pollCtx, 500)
		cancel()
		if err := fetches.Err(); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		next, err = p.apply(ctx, next, fetches.Records())
		if err != nil {
			return err
		}
	}
	return ctx.Err()
}
