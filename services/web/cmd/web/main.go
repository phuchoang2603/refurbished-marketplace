package main

import (
	"log"
	"net/http"
	"os"

	"refurbished-marketplace/services/web/internal/handlers"
	authconfig "refurbished-marketplace/shared/auth/config"
	"refurbished-marketplace/shared/proto/cartclient"
	"refurbished-marketplace/shared/proto/ordersclient"
	"refurbished-marketplace/shared/proto/paymentclient"
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

	ordersGRPCAddr := os.Getenv("ORDERS_SVC_ADDR")
	if ordersGRPCAddr == "" {
		log.Fatal("ORDERS_SVC_ADDR is required")
	}

	ordersClient, err := ordersclient.New(ordersGRPCAddr)
	if err != nil {
		log.Fatalf("orders grpc client: %v", err)
	}
	defer ordersClient.Close()

	cartGRPCAddr := os.Getenv("CART_SVC_ADDR")
	if cartGRPCAddr == "" {
		log.Fatal("CART_SVC_ADDR is required")
	}

	cartClient, err := cartclient.New(cartGRPCAddr)
	if err != nil {
		log.Fatalf("cart grpc client: %v", err)
	}
	defer cartClient.Close()

	paymentGRPCAddr := os.Getenv("PAYMENT_SVC_ADDR")
	if paymentGRPCAddr == "" {
		log.Fatal("PAYMENT_SVC_ADDR is required")
	}

	paymentClient, err := paymentclient.New(paymentGRPCAddr)
	if err != nil {
		log.Fatalf("payment grpc client: %v", err)
	}
	defer paymentClient.Close()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	h := handlers.New(usersClient, productsClient, ordersClient, cartClient, paymentClient, authconfig.DefaultConfig(jwtSecret))
	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("starting web service on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
