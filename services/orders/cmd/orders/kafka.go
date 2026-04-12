package main

import (
	"context"
	"log"
	"os"

	"refurbished-marketplace/services/orders/internal/service"
	"refurbished-marketplace/shared/messaging"
)

func runPaymentResultConsumer(ctx context.Context, svc *service.Service, bootstrap string) error {
	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "orders-service"
	}

	prefer := os.Getenv("KAFKA_PREFER_IPV4") == "1" || os.Getenv("KAFKA_PREFER_IPV4") == "true"

	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          groupID,
		Topics: []string{
			messaging.EventTypePaymentItemSucceeded,
			messaging.EventTypePaymentItemFailed,
		},
		PreferIPv4: prefer,
	}, svc.KafkaPaymentResultHandler())
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("kafka consumer close: %v", err)
		}
	}()

	log.Printf("kafka consumer started (topics payment.item.* group=%s)", groupID)
	return consumer.Run(ctx)
}
