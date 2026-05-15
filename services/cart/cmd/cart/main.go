package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"refurbished-marketplace/services/cart/internal/grpcserver"
	"refurbished-marketplace/services/cart/internal/service"

	cartv1 "refurbished-marketplace/shared/proto/cart/v1"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":9094"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR is required")
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()

	svc := service.New(rdb, 24*time.Hour)
	grpcSvc := grpcserver.New(svc)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	server := grpc.NewServer()
	cartv1.RegisterCartServiceServer(server, grpcSvc)
	reflection.Register(server)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	log.Printf("starting cart grpc service on %s", addr)
	log.Fatal(server.Serve(lis))
}
