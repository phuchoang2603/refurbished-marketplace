package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
)

type Accounts interface {
	CreatedAt(context.Context, ...string) ([]time.Time, error)
}
type Service struct {
	db       *sql.DB
	config   Config
	accounts Accounts
	gateway  *gatewayClient
}

func New(db *sql.DB, config Config, accounts Accounts) *Service {
	return &Service{db: db, config: config, accounts: accounts, gateway: newGatewayClient(config)}
}
func (service *Service) KafkaResultHandler() messaging.KafkaHandler { return service.HandleResult }
