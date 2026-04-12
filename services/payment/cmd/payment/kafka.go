package main

import (
	"context"
	"log"
	"os"
	"refurbished-marketplace/services/payment/internal/service"
	"refurbished-marketplace/shared/messaging"
)

func runOrdersItemCreatedConsumer(ctx context.Context, svc *service.Service, bootstrap string) error {
	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "payment-service"
	}

	prefer := os.Getenv("KAFKA_PREFER_IPV4") == "1" || os.Getenv("KAFKA_PREFER_IPV4") == "true"

	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          groupID,
		Topics:           []string{messaging.EventTypeOrderItemCreated},
		PreferIPv4:       prefer,
	}, svc.KafkaOrdersItemCreatedHandler())
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("kafka consumer close: %v", err)
		}
	}()

	log.Printf("kafka consumer started (topic=%s group=%s)", messaging.EventTypeOrderItemCreated, groupID)
	return consumer.Run(ctx)
}
