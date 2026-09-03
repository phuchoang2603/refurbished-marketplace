package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/phuchoang2603/refurbished-marketplace/services/users/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/users/internal/service"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	"github.com/phuchoang2603/refurbished-marketplace/shared/runtime"

	usersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/users/v1"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	runtime.InitLogging("users")
	addr := runtime.EnvOr("GRPC_ADDR", ":9091")

	db, err := runtime.OpenPostgres(runtime.MustEnv("DB_URL"))
	if err != nil {
		sharedlog.Fatal("open postgres", "err", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			sharedlog.Error("close db", "err", err)
		}
	}()

	cfg := service.DefaultConfig(os.Getenv("JWT_SECRET"))
	if err := service.ValidateConfig(cfg); err != nil {
		sharedlog.Fatal("auth config", "err", err)
	}

	svc := service.New(db, cfg)
	grpcSvc := grpcserver.New(svc)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := runtime.InitTracing(ctx, "users")
	if err != nil {
		sharedlog.Fatal("init tracing", "err", err)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			sharedlog.Error("tracing shutdown", "err", err)
		}
	}()

	shutdownMetrics, err := runtime.InitMetrics(ctx, "users")
	if err != nil {
		sharedlog.Fatal("init metrics", "err", err)
	}
	defer func() {
		if err := shutdownMetrics(context.Background()); err != nil {
			sharedlog.Error("metrics shutdown", "err", err)
		}
	}()

	if err := runtime.ServeGRPC(ctx, runtime.GRPCServerConfig{
		Addr:        addr,
		ServiceName: "users",
		Register: func(server *grpc.Server) {
			usersv1.RegisterUsersServiceServer(server, grpcSvc)
		},
	}); err != nil {
		sharedlog.Fatal("grpc serve", "err", err)
	}
}
