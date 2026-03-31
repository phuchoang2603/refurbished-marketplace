package main

import (
	"log"
	"net/http"
	"os"

	"refurbished-marketplace/services/web/internal/handlers"
	"refurbished-marketplace/shared/proto/productsclient"
	"refurbished-marketplace/shared/proto/usersclient"
)

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	usersGRPCAddr := os.Getenv("USERS_SVC_ADDR")
	if usersGRPCAddr == "" {
		log.Fatal("USERS_SVC_ADDR is required")
	}

	usersClient, err := usersclient.New(usersGRPCAddr)
	if err != nil {
		log.Fatalf("users grpc client: %v", err)
	}
	defer usersClient.Close()

	productsGRPCAddr := os.Getenv("PRODUCTS_SVC_ADDR")
	if productsGRPCAddr == "" {
		log.Fatal("PRODUCTS_SVC_ADDR is required")
	}

	productsClient, err := productsclient.New(productsGRPCAddr)
	if err != nil {
		log.Fatalf("products grpc client: %v", err)
	}
	defer productsClient.Close()

	h := handlers.New(usersClient, productsClient)
	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("starting web service on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
