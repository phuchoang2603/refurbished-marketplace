package main

import (
	"context"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"

	"github.com/phuchoang2603/refurbished-marketplace/services/inventory/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
)

func runReservationConsumer(ctx context.Context, svc *service.Service, bootstrap []string, groupID string) error {
	return runInventoryConsumer(ctx, messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          groupID,
		Topics: []string{
			messaging.EventTypeOrderCreated,
			messaging.EventTypePaymentSucceeded,
			messaging.EventTypePaymentFailed,
		},
		TracerName: "inventory",
	}, svc.KafkaReservationHandler(), "orders.created,payment.*")
}

func runProductCreatedConsumer(ctx context.Context, svc *service.Service, bootstrap []string, groupID string) error {
	return runInventoryConsumer(ctx, messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          groupID,
		Topics:           []string{messaging.EventTypeProductCreated},
		TracerName:       "inventory",
	}, svc.KafkaProductCreatedHandler(), messaging.EventTypeProductCreated)
}

func runInventoryConsumer(ctx context.Context, cfg messaging.KafkaConsumerConfig, handler messaging.KafkaHandler, topics string) error {
	consumer, err := messaging.NewKafkaConsumer(cfg, handler)
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			sharedlog.Error("kafka consumer close", "err", err)
		}
	}()

	sharedlog.Info("kafka consumer started", "topics", topics, "group", cfg.GroupID)
	return consumer.Run(ctx)
}
