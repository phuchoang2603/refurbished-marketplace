package main

import (
	"context"

	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
)

var resultTopics = []string{
	messaging.EventTypeCheckoutOrderCreated,
	messaging.EventTypeCheckoutOrderRejected,
	messaging.EventTypeCheckoutStockReserved,
	messaging.EventTypeCheckoutStockRejected,
	messaging.EventTypeCheckoutStockReleased,
	messaging.EventTypeCheckoutStockCommitted,
	messaging.EventTypeCheckoutSessionReady,
	messaging.EventTypeCheckoutSessionRejected,
	messaging.EventTypeCheckoutPaymentCancelled,
	messaging.EventTypeCheckoutPaymentSucceeded,
	messaging.EventTypeCheckoutPaymentFailed,
	messaging.EventTypeCheckoutPaymentUncertain,
	messaging.EventTypeCheckoutOrderFinalized,
}

func runCheckoutResultConsumer(ctx context.Context, handler messaging.KafkaHandler, brokers []string, groupID string) error {
	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: brokers,
		GroupID:          groupID,
		Topics:           resultTopics,
		TracerName:       "checkout",
	}, handler)
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			sharedlog.Error("kafka consumer close", "err", err)
		}
	}()
	return consumer.Run(ctx)
}
