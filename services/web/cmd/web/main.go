package main

import (
	"log"
	"net/http"
	"os"

	"refurbished-marketplace/services/web/internal/handlers"
	authconfig "refurbished-marketplace/shared/auth/config"
	cartproto "refurbished-marketplace/shared/proto/cart"
	ordersproto "refurbished-marketplace/shared/proto/orders"
	paymentproto "refurbished-marketplace/shared/proto/payment"
	productsproto "refurbished-marketplace/shared/proto/products"
	usersproto "refurbished-marketplace/shared/proto/users"
)

func requiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	usersClient, err := usersproto.New(requiredEnv("USERS_SVC_ADDR"))
	if err != nil {
		log.Fatalf("users grpc client: %v", err)
	}
	defer usersClient.Close()

	productsClient, err := productsproto.New(requiredEnv("PRODUCTS_SVC_ADDR"))
	if err != nil {
		log.Fatalf("products grpc client: %v", err)
	}
	defer productsClient.Close()

	ordersClient, err := ordersproto.New(requiredEnv("ORDERS_SVC_ADDR"))
	if err != nil {
		log.Fatalf("orders grpc client: %v", err)
	}
	defer ordersClient.Close()

	cartClient, err := cartproto.New(requiredEnv("CART_SVC_ADDR"))
	if err != nil {
		log.Fatalf("cart grpc client: %v", err)
	}
	defer cartClient.Close()

	paymentClient, err := paymentproto.New(requiredEnv("PAYMENT_SVC_ADDR"))
	if err != nil {
		log.Fatalf("payment grpc client: %v", err)
	}
	defer paymentClient.Close()

	jwtSecret := requiredEnv("JWT_SECRET")

	h := handlers.New(usersClient, productsClient, ordersClient, cartClient, paymentClient, authconfig.DefaultConfig(jwtSecret))
	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("starting web service on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
