package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	testkafka "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/kafka"
	testredis "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/redis"
	"github.com/twmb/franz-go/pkg/kgo"
)

func accountRecord(t *testing.T, id string, offset int64) *kgo.Record {
	t.Helper()
	payload, err := json.Marshal(messaging.AccountCreated{UserID: id, CreatedAt: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	return &kgo.Record{Topic: messaging.EventTypeAccountCreated, Key: []byte(id), Offset: offset, Value: payload}
}

func TestIntegrationProjectionCheckpointRecovery(t *testing.T) {
	client := testredis.SetupRedisContainer(t)
	p := New(client)
	id := uuid.NewString()
	record := accountRecord(t, id, 4)
	next, err := p.apply(t.Context(), 0, []*kgo.Record{record})
	if err != nil || next != 5 {
		t.Fatalf("next=%d err=%v", next, err)
	}
	restarted := New(client)
	saved, err := restarted.nextOffset(t.Context())
	if err != nil || saved != next {
		t.Fatalf("saved=%d err=%v", saved, err)
	}
	times, err := restarted.CreatedAt(t.Context(), id, id)
	if err != nil || len(times) != 2 || !times[0].Equal(times[1]) {
		t.Fatalf("times=%v err=%v", times, err)
	}
	if _, err := p.apply(t.Context(), 0, []*kgo.Record{record}); !errors.Is(err, ErrCheckpointChanged) {
		t.Fatalf("stale writer: %v", err)
	}
	if err := client.FlushDB(t.Context()).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := p.apply(t.Context(), 5, []*kgo.Record{accountRecord(t, uuid.NewString(), 5)}); !errors.Is(err, ErrCheckpointChanged) {
		t.Fatalf("lost projection must reject advanced writer: %v", err)
	}
	saved, err = restarted.nextOffset(t.Context())
	if err != nil || saved != 0 {
		t.Fatalf("rebuild offset=%d err=%v", saved, err)
	}
	if _, err := restarted.apply(t.Context(), saved, []*kgo.Record{record}); err != nil {
		t.Fatal(err)
	}
	if _, err := restarted.CreatedAt(t.Context(), id); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationInvalidBatchIsAtomic(t *testing.T) {
	client := testredis.SetupRedisContainer(t)
	p := New(client)
	id := uuid.NewString()
	malformed := accountRecord(t, uuid.NewString(), 1)
	malformed.Value = []byte(`{`)
	if _, err := p.apply(t.Context(), 0, []*kgo.Record{accountRecord(t, id, 0), malformed}); err == nil {
		t.Fatal("accepted invalid batch")
	}
	if size := client.DBSize(t.Context()).Val(); size != 0 {
		t.Fatalf("invalid batch wrote %d keys", size)
	}
}

func TestIntegrationAccountConsumerReplaysAfterRedisLoss(t *testing.T) {
	client := testredis.SetupRedisContainer(t)
	kafka := testkafka.SetupKafka(t)
	brokers, err := kafka.Brokers(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.NewString()
	record := accountRecord(t, id, 0)
	testkafka.ProduceKafkaRecord(t, t.Context(), brokers, record.Topic, record.Value)
	p := New(client)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- p.Run(ctx, brokers) }()
	waitForAccount(t, p, id, done)
	if err := client.FlushDB(t.Context()).Err(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if !errors.Is(err, ErrCheckpointChanged) {
			t.Fatalf("expected recovery restart, got %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("consumer did not detect Redis loss while idle")
	}
	go func() { done <- p.Run(ctx, brokers) }()
	waitForAccount(t, p, id, done)
	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("consumer did not stop")
	}
}

func waitForAccount(t *testing.T, p *Projection, id string, done <-chan error) {
	t.Helper()
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-done:
			t.Fatalf("consumer stopped: %v", err)
		case <-timer.C:
			t.Fatal("account was not projected")
		case <-ticker.C:
			if _, err := p.CreatedAt(t.Context(), id); err == nil {
				return
			}
		}
	}
}
