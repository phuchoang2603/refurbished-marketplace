package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"refurbished-marketplace/services/inventory/internal/database"
	"refurbished-marketplace/services/inventory/internal/grpcserver"
	"refurbished-marketplace/services/inventory/internal/service"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	inventoryv1 "refurbished-marketplace/shared/proto/inventory/v1"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":9095"
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
	inventoryv1.RegisterInventoryServiceServer(server, grpcSvc)
	reflection.Register(server)

	log.Printf("starting inventory grpc service on %s", addr)
	log.Fatal(server.Serve(lis))
}
