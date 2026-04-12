package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"refurbished-marketplace/services/payment/internal/database"
	"refurbished-marketplace/services/payment/internal/grpcserver"
	"refurbished-marketplace/services/payment/internal/service"
	"sync"
	"syscall"

	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":9096"
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

	queries := database.New(db)
	svc := service.New(queries, db)
	grpcSvc := grpcserver.New(svc)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	server := grpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(server, grpcSvc)
	reflection.Register(server)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	if kafkaBootstrap := os.Getenv("KAFKA_BOOTSTRAP_SERVERS"); kafkaBootstrap != "" {
		wg.Go(func() {
			if err := runOrdersItemCreatedConsumer(ctx, svc, kafkaBootstrap); err != nil && !errors.Is(err, context.Canceled) {
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

	log.Printf("starting payment grpc service on %s", addr)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("grpc: %v", err)
	}
	wg.Wait()
}
