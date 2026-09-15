package main

import (
	"context"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"

	"github.com/phuchoang2603/refurbished-marketplace/services/search/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
)

func runProductCreatedConsumer(ctx context.Context, svc *service.Service, bootstrap []string, groupID string) error {
	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          groupID,
		Topics:           []string{messaging.EventTypeProductCreated},
		TracerName:       "search",
	}, svc.KafkaProductCreatedHandler())
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			sharedlog.Error("kafka consumer close", "err", err)
		}
	}()

	sharedlog.Info("kafka consumer started", "topics", messaging.EventTypeProductCreated, "group", groupID)
	return consumer.Run(ctx)
}
