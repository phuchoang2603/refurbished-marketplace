package main

import (
	"context"
	"log"
	"os"

	"refurbished-marketplace/services/orders/internal/service"
	"refurbished-marketplace/shared/messaging"
)

func runOrderResultConsumer(ctx context.Context, svc *service.Service, bootstrap []string) error {
	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "orders-service"
	}

	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          groupID,
		Topics: []string{
			messaging.EventTypeInventoryReservationFailed,
			messaging.EventTypePaymentSucceeded,
			messaging.EventTypePaymentFailed,
		},
	}, svc.KafkaOrderResultHandler())
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("kafka consumer close: %v", err)
		}
	}()

	log.Printf("kafka consumer started (topics inventory/payment results group=%s)", groupID)
	return consumer.Run(ctx)
}
