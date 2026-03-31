package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"refurbished-marketplace/services/products/internal/database"
	"refurbished-marketplace/services/products/internal/grpcserver"
	"refurbished-marketplace/services/products/internal/service"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"
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

	queries := database.New(db)
	svc := service.New(queries)
	grpcSvc := grpcserver.New(svc)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	server := grpc.NewServer()
	productsv1.RegisterProductsServiceServer(server, grpcSvc)
	reflection.Register(server)

	log.Printf("starting products grpc service on %s", addr)
	log.Fatal(server.Serve(lis))
}
