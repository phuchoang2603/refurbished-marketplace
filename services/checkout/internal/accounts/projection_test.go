package accounts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
)

type stubStore struct {
	values    []interface{}
	offset    string
	err       error
	evalCalls int
	applied   interface{}
}

func (s *stubStore) HMGet(context.Context, string, ...string) *redis.SliceCmd {
	return redis.NewSliceResult(s.values, s.err)
}

func (s *stubStore) HGet(context.Context, string, string) *redis.StringCmd {
	return redis.NewStringResult(s.offset, s.err)
}

func (s *stubStore) Eval(context.Context, string, []string, ...interface{}) *redis.Cmd {
	s.evalCalls++
	return redis.NewCmdResult(s.applied, s.err)
}

func TestReadAccountFacts(t *testing.T) {
	for _, tc := range []struct {
		name      string
		values    []interface{}
		err       error
		wantError bool
	}{
		{name: "present", values: []interface{}{"2020-01-02T03:04:05+02:00", "2021-02-03T04:05:06Z"}},
		{name: "missing", values: []interface{}{nil, "2021-02-03T04:05:06Z"}, wantError: true},
		{name: "invalid", values: []interface{}{"invalid", "2021-02-03T04:05:06Z"}, wantError: true},
		{name: "zero", values: []interface{}{"0001-01-01T00:00:00Z", "2021-02-03T04:05:06Z"}, wantError: true},
		{name: "unavailable", err: errors.New("redis unavailable"), wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &Projection{store: &stubStore{values: tc.values, err: tc.err}}
			times, err := p.CreatedAt(t.Context(), "seller", "buyer")
			if (err != nil) != tc.wantError {
				t.Fatalf("times=%v err=%v", times, err)
			}
			if !tc.wantError && (len(times) != 2 || times[0].Hour() != 1 || times[0].Location() != time.UTC) {
				t.Fatalf("times=%v", times)
			}
		})
	}
}

func TestInvalidBatchDoesNotAdvanceCheckpoint(t *testing.T) {
	for _, payload := range []string{`{`, `{"user_id":"invalid","created_at":"2020-01-01T00:00:00Z"}`, `{"user_id":"00000000-0000-0000-0000-000000000000","created_at":"2020-01-01T00:00:00Z"}`, `{"user_id":"6eab0af2-59f8-4ff4-8c7b-e77d22fde9b9"}`} {
		store := &stubStore{}
		p := &Projection{store: store}
		next, err := p.apply(t.Context(), 7, []*kgo.Record{{Topic: messaging.EventTypeAccountCreated, Offset: 9, Value: []byte(payload)}})
		if err == nil || next != 7 || store.evalCalls != 0 {
			t.Fatalf("next=%d err=%v writes=%d", next, err, store.evalCalls)
		}
	}
}

func TestCheckpointErrors(t *testing.T) {
	for _, tc := range []struct {
		offset    string
		err       error
		want      int64
		wantError bool
	}{
		{err: redis.Nil}, {offset: "42", want: 42}, {offset: "-1", wantError: true}, {offset: "garbage", wantError: true}, {err: errors.New("redis unavailable"), wantError: true},
	} {
		p := &Projection{store: &stubStore{offset: tc.offset, err: tc.err}}
		next, err := p.nextOffset(t.Context())
		if next != tc.want || (err != nil) != tc.wantError {
			t.Fatalf("next=%d err=%v", next, err)
		}
	}
}
