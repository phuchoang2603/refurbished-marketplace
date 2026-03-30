package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"refurbished-marketplace/services/users/internal/database"
	"refurbished-marketplace/services/users/internal/grpcserver"
	"refurbished-marketplace/services/users/internal/service"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	usersv1 "refurbished-marketplace/services/users/proto/v1"
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
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	queries := database.New(db)
	jwtSecret := os.Getenv("JWT_SECRET")
	cfg := service.DefaultConfig(jwtSecret)
	if err := service.ValidateConfig(cfg); err != nil {
		log.Fatalf("auth config: %v", err)
	}

	svc := service.New(queries, cfg)
	grpcSvc := grpcserver.New(svc)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	server := grpc.NewServer()
	usersv1.RegisterUsersServiceServer(server, grpcSvc)
	reflection.Register(server)

	log.Printf("starting users grpc service on %s", addr)
	log.Fatal(server.Serve(lis))
}
