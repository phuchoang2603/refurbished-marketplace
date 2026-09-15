package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/meilisearch/meilisearch-go"
	"github.com/phuchoang2603/refurbished-marketplace/services/search/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/search/internal/service"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	searchv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/search/v1"
	"github.com/phuchoang2603/refurbished-marketplace/shared/runtime"

	"google.golang.org/grpc"
)

func main() {
	runtime.InitLogging("search")
	cfg := service.LoadConfig()
	if err := service.ValidateConfig(cfg); err != nil {
		sharedlog.Fatal("config", "err", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := runtime.InitTracing(ctx, "search")
	if err != nil {
		sharedlog.Fatal("init tracing", "err", err)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			sharedlog.Error("tracing shutdown", "err", err)
		}
	}()

	shutdownMetrics, err := runtime.InitMetrics(ctx, "search")
	if err != nil {
		sharedlog.Fatal("init metrics", "err", err)
	}
	defer func() {
		if err := shutdownMetrics(context.Background()); err != nil {
			sharedlog.Error("metrics shutdown", "err", err)
		}
	}()

	client := meilisearch.New(cfg.MeiliURL, meilisearch.WithAPIKey(cfg.MeiliAPIKey))
	svc := service.New(client, cfg.ListingsIndex)
	if err := svc.EnsureIndexSettings(ctx); err != nil {
		sharedlog.Fatal("ensure index settings", "err", err)
	}
	grpcSvc := grpcserver.New(svc)

	var wg sync.WaitGroup
	runtime.StartKafkaConsumer(ctx, &wg, func(ctx context.Context, brokers []string) error {
		return runProductCreatedConsumer(ctx, svc, brokers, cfg.ProductCreatedKafkaGroupID)
	})

	if err := runtime.ServeGRPC(ctx, runtime.GRPCServerConfig{
		Addr:        cfg.GRPCAddr,
		ServiceName: "search",
		Register: func(server *grpc.Server) {
			searchv1.RegisterSearchServiceServer(server, grpcSvc)
		},
	}); err != nil {
		sharedlog.Fatal("grpc serve", "err", err)
	}
	wg.Wait()
}
