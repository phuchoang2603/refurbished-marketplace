package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"refurbished-marketplace/services/products/internal/grpcserver"
	"refurbished-marketplace/services/products/internal/service"
	"refurbished-marketplace/shared/messaging"

	productsv1 "refurbished-marketplace/shared/proto/products/v1"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":9092"
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
	productsv1.RegisterProductsServiceServer(server, grpcSvc)
	reflection.Register(server)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	if raw := os.Getenv("KAFKA_BOOTSTRAP_SERVERS"); raw != "" {
		brokers := messaging.ParseBootstrapServers(raw)
		if len(brokers) == 0 {
			log.Print("KAFKA_BOOTSTRAP_SERVERS has no brokers after parsing; skipping Kafka consumer")
		} else {
			go func() {
				if err := runReservationConsumer(ctx, svc, brokers); err != nil && !errors.Is(err, context.Canceled) {
					log.Printf("kafka consumer: %v", err)
				}
			}()
		}
	} else {
		log.Print("KAFKA_BOOTSTRAP_SERVERS not set; skipping Kafka consumer")
	}

	log.Printf("starting products grpc service on %s", addr)
	log.Fatal(server.Serve(lis))
}
