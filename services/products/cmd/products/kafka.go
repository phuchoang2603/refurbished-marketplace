package main

import (
	"context"
	"log"

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
			log.Printf("kafka consumer close: %v", err)
		}
	}()

	log.Printf("products reservation consumer started (topics orders.created,payment.* group=%s)", groupID)
	return consumer.Run(ctx)
}
