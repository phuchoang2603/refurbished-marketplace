package service

import (
	"context"
	"database/sql"

	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	usersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/users/v1"
	"google.golang.org/grpc"
)

type Users interface {
	GetUserByID(context.Context, *usersv1.GetUserByIDRequest, ...grpc.CallOption) (*usersv1.User, error)
}
type Service struct {
	db      *sql.DB
	config  Config
	users   Users
	gateway *gatewayClient
}

func New(db *sql.DB, config Config, users Users) *Service {
	return &Service{db: db, config: config, users: users, gateway: newGatewayClient(config)}
}
func (service *Service) KafkaResultHandler() messaging.KafkaHandler { return service.HandleResult }
