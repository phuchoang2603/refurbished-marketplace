package service

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func (service *Service) RegisterMetrics() (metric.Registration, error) {
	meter := otel.Meter("checkout")
	stuck, err := meter.Int64ObservableGauge("checkout.stuck.count")
	if err != nil {
		return nil, err
	}
	pendingAge, err := meter.Int64ObservableGauge("checkout.pending.max_age.seconds")
	if err != nil {
		return nil, err
	}
	compensating, err := meter.Int64ObservableGauge("checkout.compensating.count")
	if err != nil {
		return nil, err
	}
	exceptions, err := meter.Int64ObservableGauge("checkout.financial_exceptions.count")
	if err != nil {
		return nil, err
	}
	return meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		var stuckCount int64
		if err := service.db.QueryRowContext(queryCtx, `
			SELECT count(*) FROM checkouts
			WHERE state NOT IN ('CHECKOUT_STATE_COMPLETED', 'CHECKOUT_STATE_FAILED', 'CHECKOUT_STATE_NEEDS_REVIEW')
			AND (deadline_at < NOW() OR updated_at < NOW() - INTERVAL '5 minutes')
		`).Scan(&stuckCount); err != nil {
			return err
		}
		var maxPendingAge, compensatingCount int64
		if err := service.db.QueryRowContext(queryCtx, `
			SELECT COALESCE(MAX(EXTRACT(EPOCH FROM (NOW() - created_at))::BIGINT), 0)
			FROM checkouts
			WHERE state NOT IN ('CHECKOUT_STATE_COMPLETED', 'CHECKOUT_STATE_FAILED', 'CHECKOUT_STATE_NEEDS_REVIEW')
		`).Scan(&maxPendingAge); err != nil {
			return err
		}
		if err := service.db.QueryRowContext(queryCtx, `
			SELECT count(*) FROM checkouts
			WHERE state IN ('CHECKOUT_STATE_RELEASING_STOCK', 'CHECKOUT_STATE_FINALIZING_FAILED')
		`).Scan(&compensatingCount); err != nil {
			return err
		}
		var unresolvedCount int64
		if err := service.db.QueryRowContext(queryCtx, `
			SELECT count(*) FROM checkout_exceptions WHERE resolved_at IS NULL
		`).Scan(&unresolvedCount); err != nil {
			return err
		}
		observer.ObserveInt64(stuck, stuckCount)
		observer.ObserveInt64(pendingAge, maxPendingAge)
		observer.ObserveInt64(compensating, compensatingCount)
		observer.ObserveInt64(exceptions, unresolvedCount)
		return nil
	}, stuck, pendingAge, compensating, exceptions)
}
