package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/phuchoang2603/refurbished-marketplace/services/products/db/migrations"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/catalog"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/service"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	"github.com/phuchoang2603/refurbished-marketplace/shared/runtime"

	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"

	"google.golang.org/grpc"
)

func main() {
	runtime.InitLogging("products")
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := runtime.MigratePostgres(context.Background(), migrations.Files); err != nil {
			sharedlog.Fatal("migrate postgres", "err", err)
		}
		return
	}
	cfg := service.LoadConfig()
	if err := service.ValidateConfig(cfg); err != nil {
		sharedlog.Fatal("config", "err", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := runtime.OpenPostgres()
	if err != nil {
		sharedlog.Fatal("open postgres", "err", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			sharedlog.Error("close db", "err", err)
		}
	}()

	svc := service.New(catalog.New(db))
	grpcSvc := grpcserver.New(svc)

	shutdownTracing, err := runtime.InitTracing(ctx, "products")
	if err != nil {
		sharedlog.Fatal("init tracing", "err", err)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			sharedlog.Error("tracing shutdown", "err", err)
		}
	}()

	shutdownMetrics, err := runtime.InitMetrics(ctx, "products")
	if err != nil {
		sharedlog.Fatal("init metrics", "err", err)
	}
	defer func() {
		if err := shutdownMetrics(context.Background()); err != nil {
			sharedlog.Error("metrics shutdown", "err", err)
		}
	}()

	if err := runtime.ServeGRPC(ctx, runtime.GRPCServerConfig{
		Addr:        cfg.GRPCAddr,
		ServiceName: "products",
		Register: func(server *grpc.Server) {
			productsv1.RegisterProductsServiceServer(server, grpcSvc)
		},
	}); err != nil {
		sharedlog.Fatal("grpc serve", "err", err)
	}
}
