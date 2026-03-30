package main

import (
	"log"
	"net/http"
	"os"

	"refurbished-marketplace/services/web/internal/handlers"
	"refurbished-marketplace/shared/proto/usersclient"
)

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	usersGRPCAddr := os.Getenv("USERS_GRPC_ADDR")
	if usersGRPCAddr == "" {
		log.Fatal("USERS_GRPC_ADDR is required")
	}

	usersClient, err := usersclient.New(usersGRPCAddr)
	if err != nil {
		log.Fatalf("users grpc client: %v", err)
	}
	defer usersClient.Close()

	h := handlers.New(usersClient)
	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("starting web service on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
