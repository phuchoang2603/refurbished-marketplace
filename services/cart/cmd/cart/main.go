package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"refurbished-marketplace/services/cart/internal/grpcserver"
	"refurbished-marketplace/services/cart/internal/service"
	"refurbished-marketplace/shared/runtime"

	cartv1 "refurbished-marketplace/shared/proto/cart/v1"

	"google.golang.org/grpc"
)

func main() {
	cfg := service.LoadConfig()
	if err := service.ValidateConfig(cfg); err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rdb, err := runtime.OpenRedis(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Printf("close redis: %v", err)
		}
	}()

	svc := service.New(rdb, cfg)
	grpcSvc := grpcserver.New(svc)

	if err := runtime.ServeGRPC(ctx, runtime.GRPCServerConfig{
		Addr:        cfg.GRPCAddr,
		ServiceName: "cart",
		Register: func(server *grpc.Server) {
			cartv1.RegisterCartServiceServer(server, grpcSvc)
		},
	}); err != nil {
		log.Fatalf("grpc: %v", err)
	}
}
