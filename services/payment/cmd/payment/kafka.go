package main

import (
	"context"
	"log/slog"

	"refurbished-marketplace/services/payment/internal/service"
	"refurbished-marketplace/shared/messaging"
)

func runInventoryReservedConsumer(ctx context.Context, svc *service.Service, bootstrap []string, groupID string) error {
	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          groupID,
		Topics:           []string{messaging.EventTypeInventoryReserved},
		TracerName:       "payment",
	}, svc.KafkaInventoryReservedHandler())
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
		"topic", messaging.EventTypeInventoryReserved,
		"group", groupID,
	)
	return consumer.Run(ctx)
}
