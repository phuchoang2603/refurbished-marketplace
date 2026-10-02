package service

import (
	"context"

	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
)

func (s *Service) KafkaProductCreatedHandler() messaging.KafkaHandler {
	return func(ctx context.Context, msg messaging.KafkaMessage) error {
		if msg.Topic != messaging.EventTypeProductCreated {
			return nil
		}
		return s.HandleProductCreated(ctx, msg.Value)
	}
}
