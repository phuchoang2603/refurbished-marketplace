package main

import (
	"context"
	"log/slog"

	"refurbished-marketplace/services/orders/internal/service"
	"refurbished-marketplace/shared/messaging"
)

func runOrderResultConsumer(ctx context.Context, svc *service.Service, bootstrap []string, groupID string) error {
	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          groupID,
		Topics: []string{
			messaging.EventTypeInventoryReservationFailed,
			messaging.EventTypePaymentSucceeded,
			messaging.EventTypePaymentFailed,
		},
		TracerName: "orders",
	}, svc.KafkaOrderResultHandler())
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			slog.Error("kafka consumer close", "err", err)
		}
	}()

	slog.Info(
		"kafka consumer started",
		"topics", "inventory/payment results",
		"group", groupID,
	)
	return consumer.Run(ctx)
}
