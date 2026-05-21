package main

import (
	"context"
	"log"
	"os"

	"refurbished-marketplace/services/products/internal/service"
	"refurbished-marketplace/shared/messaging"
)

func runReservationConsumer(ctx context.Context, svc *service.Service, bootstrap []string) error {
	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "products-service"
	}

	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          groupID,
		Topics: []string{
			messaging.EventTypeOrderCreated,
			messaging.EventTypePaymentSucceeded,
			messaging.EventTypePaymentFailed,
		},
	}, svc.KafkaReservationHandler())
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("kafka consumer close: %v", err)
		}
	}()

	log.Printf("catalog reservation consumer started (topics orders.created,payment.* group=%s)", groupID)
	return consumer.Run(ctx)
}
