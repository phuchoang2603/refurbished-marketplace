package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/service"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	"github.com/phuchoang2603/refurbished-marketplace/shared/runtime"

	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	runtime.InitLogging("products")
	cfg := service.LoadConfig()
	if err := service.ValidateConfig(cfg); err != nil {
		sharedlog.Fatal("config", "err", err)
	}

	db, err := runtime.OpenPostgres(runtime.MustEnv("DB_URL"))
	if err != nil {
		sharedlog.Fatal("open postgres", "err", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			sharedlog.Error("close db", "err", err)
		}
	}()

	svc := service.New(db)
	grpcSvc := grpcserver.New(svc)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := runtime.InitTracing(ctx, "products")
	if err != nil {
		sharedlog.Fatal("init tracing", "err", err)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			sharedlog.Error("tracing shutdown", "err", err)
		}
	}()

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
		sharedlog.Fatal("grpc serve", "err", err)
	}
	wg.Wait()
}
