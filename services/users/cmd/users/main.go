package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"refurbished-marketplace/services/users/internal/grpcserver"
	"refurbished-marketplace/services/users/internal/service"

	usersv1 "refurbished-marketplace/shared/proto/users/v1"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":9091"
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close db: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	cfg := service.DefaultConfig(jwtSecret)
	if err := service.ValidateConfig(cfg); err != nil {
		log.Fatalf("auth config: %v", err)
	}

	svc := service.New(db, cfg)
	grpcSvc := grpcserver.New(svc)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	server := grpc.NewServer()
	usersv1.RegisterUsersServiceServer(server, grpcSvc)
	reflection.Register(server)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	log.Printf("starting users grpc service on %s", addr)
	log.Fatal(server.Serve(lis))
}
