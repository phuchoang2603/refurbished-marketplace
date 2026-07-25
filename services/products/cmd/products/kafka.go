package main

import (
	"context"
	"log/slog"

	"refurbished-marketplace/services/products/internal/service"
	"refurbished-marketplace/shared/messaging"
)

func runReservationConsumer(ctx context.Context, svc *service.Service, bootstrap []string, groupID string) error {
	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          groupID,
		Topics: []string{
			messaging.EventTypeOrderCreated,
			messaging.EventTypePaymentSucceeded,
			messaging.EventTypePaymentFailed,
		},
		TracerName: "products",
	}, svc.KafkaReservationHandler())
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
		"topics", "orders.created,payment.*",
		"group", groupID,
	)
	return consumer.Run(ctx)
}
