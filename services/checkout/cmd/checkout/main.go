package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/service"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"github.com/phuchoang2603/refurbished-marketplace/shared/runtime"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	runtime.InitLogging("checkout")
	config := service.LoadConfig()
	if err := service.ValidateConfig(config); err != nil {
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := runtime.InitTracing(ctx, "checkout")
	if err != nil {
		sharedlog.Fatal("init tracing", "err", err)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			sharedlog.Error("tracing shutdown", "err", err)
		}
	}()
	shutdownMetrics, err := runtime.InitMetrics(ctx, "checkout")
	if err != nil {
		sharedlog.Fatal("init metrics", "err", err)
	}
	defer func() {
		if err := shutdownMetrics(context.Background()); err != nil {
			sharedlog.Error("metrics shutdown", "err", err)
		}
	}()

	checkoutService := service.New(db)
	metricRegistration, err := checkoutService.RegisterMetrics()
	if err != nil {
		sharedlog.Fatal("checkout metrics", "err", err)
	}
	defer func() {
		if err := metricRegistration.Unregister(); err != nil {
			sharedlog.Error("unregister checkout metrics", "err", err)
		}
	}()
	var workers sync.WaitGroup
	runtime.StartKafkaConsumer(ctx, &workers, func(ctx context.Context, brokers []string) error {
		return runCheckoutResultConsumer(ctx, checkoutService.KafkaResultHandler(), brokers, config.KafkaGroupID)
	})
	workers.Go(func() {
		if err := checkoutService.RunDeadlines(ctx); err != nil && ctx.Err() == nil {
			sharedlog.Error("checkout deadline worker stopped", "err", err)
		}
	})
	if err := runtime.ServeGRPC(ctx, runtime.GRPCServerConfig{
		Addr:        config.GRPCAddr,
		ServiceName: "checkout",
		Register: func(server *grpc.Server) {
			checkoutv1.RegisterCheckoutServiceServer(server, grpcserver.New(checkoutService))
		},
	}); err != nil {
		sharedlog.Fatal("grpc serve", "err", err)
	}
	workers.Wait()
}
