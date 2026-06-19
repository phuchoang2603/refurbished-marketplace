package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"refurbished-marketplace/services/products/internal/grpcserver"
	"refurbished-marketplace/services/products/internal/service"
	"refurbished-marketplace/shared/runtime"

	productsv1 "refurbished-marketplace/shared/proto/products/v1"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	cfg := service.LoadConfig()
	if err := service.ValidateConfig(cfg); err != nil {
		log.Fatal(err)
	}

	db, err := runtime.OpenPostgres(runtime.MustEnv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close db: %v", err)
		}
	}()

	svc := service.New(db)
	grpcSvc := grpcserver.New(svc)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	runtime.StartKafkaConsumer(ctx, &wg, func(ctx context.Context, brokers []string) error {
		return runReservationConsumer(ctx, svc, brokers, cfg.KafkaGroupID)
	})

	if err := runtime.ServeGRPC(ctx, runtime.GRPCServerConfig{
		Addr:        cfg.GRPCAddr,
		ServiceName: "products",
		Register: func(server *grpc.Server) {
			productsv1.RegisterProductsServiceServer(server, grpcSvc)
		},
	}); err != nil {
		log.Fatalf("grpc: %v", err)
	}
	wg.Wait()
}
