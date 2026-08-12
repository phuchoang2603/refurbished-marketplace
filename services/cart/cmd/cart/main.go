package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/phuchoang2603/refurbished-marketplace/services/cart/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/cart/internal/service"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	"github.com/phuchoang2603/refurbished-marketplace/shared/runtime"

	cartv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/cart/v1"

	"google.golang.org/grpc"
)

func main() {
	runtime.InitLogging("cart")
	cfg := service.LoadConfig()
	if err := service.ValidateConfig(cfg); err != nil {
		sharedlog.Fatal("config", "err", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := runtime.InitTracing(ctx, "cart")
	if err != nil {
		sharedlog.Fatal("init tracing", "err", err)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			sharedlog.Error("tracing shutdown", "err", err)
		}
	}()

	rdb, err := runtime.OpenRedis(ctx, cfg.RedisAddr)
	if err != nil {
		sharedlog.Fatal("open redis", "err", err)
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			sharedlog.Error("close redis", "err", err)
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
		sharedlog.Fatal("grpc serve", "err", err)
	}
}
