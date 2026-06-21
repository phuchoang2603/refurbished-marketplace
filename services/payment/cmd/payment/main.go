package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"refurbished-marketplace/services/payment/internal/grpcserver"
	"refurbished-marketplace/services/payment/internal/service"
	"refurbished-marketplace/shared/runtime"

	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

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
		return runInventoryReservedConsumer(ctx, svc, brokers, cfg.KafkaGroupID)
	})

	if err := runtime.ServeGRPC(ctx, runtime.GRPCServerConfig{
		Addr:        cfg.GRPCAddr,
		ServiceName: "payment",
		Register: func(server *grpc.Server) {
			paymentv1.RegisterPaymentServiceServer(server, grpcSvc)
		},
	}); err != nil {
		log.Fatalf("grpc: %v", err)
	}
	wg.Wait()
}
