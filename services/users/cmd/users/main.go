package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"refurbished-marketplace/services/users/internal/grpcserver"
	"refurbished-marketplace/services/users/internal/service"
	"refurbished-marketplace/shared/runtime"

	usersv1 "refurbished-marketplace/shared/proto/users/v1"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	addr := runtime.EnvOr("GRPC_ADDR", ":9091")

	db, err := runtime.OpenPostgres(runtime.MustEnv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close db: %v", err)
		}
	}()

	cfg := service.DefaultConfig(os.Getenv("JWT_SECRET"))
	if err := service.ValidateConfig(cfg); err != nil {
		log.Fatalf("auth config: %v", err)
	}

	svc := service.New(db, cfg)
	grpcSvc := grpcserver.New(svc)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := runtime.InitTracing(ctx, "users")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			log.Printf("tracing shutdown: %v", err)
		}
	}()

	if err := runtime.ServeGRPC(ctx, runtime.GRPCServerConfig{
		Addr:        addr,
		ServiceName: "users",
		Register: func(server *grpc.Server) {
			usersv1.RegisterUsersServiceServer(server, grpcSvc)
		},
	}); err != nil {
		log.Fatalf("grpc: %v", err)
	}
}
