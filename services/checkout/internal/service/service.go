package service

import (
	"database/sql"

	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
)

type Service struct {
	db *sql.DB
}

func New(db *sql.DB) *Service {
	return &Service{db: db}
}

func (service *Service) KafkaResultHandler() messaging.KafkaHandler {
	return service.HandleResult
}
