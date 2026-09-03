package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/phuchoang2603/refurbished-marketplace/services/orders/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/orders/internal/service"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	"github.com/phuchoang2603/refurbished-marketplace/shared/runtime"

	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	runtime.InitLogging("orders")
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

	shutdownTracing, err := runtime.InitTracing(ctx, "orders")
	if err != nil {
		sharedlog.Fatal("init tracing", "err", err)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			sharedlog.Error("tracing shutdown", "err", err)
		}
	}()

	shutdownMetrics, err := runtime.InitMetrics(ctx, "orders")
	if err != nil {
		sharedlog.Fatal("init metrics", "err", err)
	}
	defer func() {
		if err := shutdownMetrics(context.Background()); err != nil {
			sharedlog.Error("metrics shutdown", "err", err)
		}
	}()

	var wg sync.WaitGroup
	runtime.StartKafkaConsumer(ctx, &wg, func(ctx context.Context, brokers []string) error {
		return runOrderResultConsumer(ctx, svc, brokers, cfg.KafkaGroupID)
	})

	if err := runtime.ServeGRPC(ctx, runtime.GRPCServerConfig{
		Addr:        cfg.GRPCAddr,
		ServiceName: "orders",
		Register: func(server *grpc.Server) {
			ordersv1.RegisterOrdersServiceServer(server, grpcSvc)
		},
	}); err != nil {
		sharedlog.Fatal("grpc serve", "err", err)
	}
	wg.Wait()
}
