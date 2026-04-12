package service

import (
	"context"

	"refurbished-marketplace/shared/messaging"
)

// KafkaOrdersItemCreatedHandler returns a handler for the orders.item.created topic.
func (s *Service) KafkaOrdersItemCreatedHandler() messaging.KafkaHandler {
	return func(ctx context.Context, msg messaging.KafkaMessage) error {
		return s.HandleOrdersItemCreated(ctx, messaging.KafkaMessageID(msg), msg.Value)
	}
}
