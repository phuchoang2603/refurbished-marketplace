package main

import (
	"context"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"

	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
)

func runCheckoutCommandConsumer(ctx context.Context, svc *service.Service, bootstrap []string, groupID string) error {
	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          groupID,
		Topics:           []string{messaging.EventTypeCheckoutCreateSession, messaging.EventTypeCheckoutCancelSession},
		TracerName:       "payment",
	}, svc.HandleCheckoutCommand)
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			sharedlog.Error("kafka consumer close", "err", err)
		}
	}()

	sharedlog.Info(
		"kafka consumer started",
		"topics", "checkout payment commands",
		"group", groupID,
	)
	return consumer.Run(ctx)
}
