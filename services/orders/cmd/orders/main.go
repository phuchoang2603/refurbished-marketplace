package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"refurbished-marketplace/services/orders/internal/grpcserver"
	"refurbished-marketplace/services/orders/internal/service"
	"sync"
	"syscall"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":9093"
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	svc := service.New(db)
	grpcSvc := grpcserver.New(svc)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	server := grpc.NewServer()
	ordersv1.RegisterOrdersServiceServer(server, grpcSvc)
	reflection.Register(server)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	if kafkaBootstrap := os.Getenv("KAFKA_BOOTSTRAP_SERVERS"); kafkaBootstrap != "" {
		wg.Go(func() {
			if err := runPaymentResultConsumer(ctx, svc, kafkaBootstrap); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("kafka consumer: %v", err)
			}
		})
	} else {
		log.Print("KAFKA_BOOTSTRAP_SERVERS not set; skipping Kafka consumer")
	}

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	log.Printf("starting orders grpc service on %s", addr)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("grpc: %v", err)
	}
	wg.Wait()
}
