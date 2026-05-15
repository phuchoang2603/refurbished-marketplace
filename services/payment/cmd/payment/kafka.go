package main

import (
	"context"
	"log"
	"os"

	"refurbished-marketplace/services/payment/internal/service"
	"refurbished-marketplace/shared/messaging"
)

func runOrdersCreatedConsumer(ctx context.Context, svc *service.Service, bootstrap []string) error {
	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "payment-service"
	}

	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          groupID,
		Topics:           []string{messaging.EventTypeOrderCreated},
	}, svc.KafkaOrdersCreatedHandler())
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("kafka consumer close: %v", err)
		}
	}()

	log.Printf("kafka consumer started (topic=%s group=%s)", messaging.EventTypeOrderCreated, groupID)
	return consumer.Run(ctx)
}
