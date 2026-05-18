package main

import (
	"log"
	"net/http"

	webclients "refurbished-marketplace/services/web/internal/clients"
	"refurbished-marketplace/services/web/internal/handlers"
	authconfig "refurbished-marketplace/shared/auth/config"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	deps, err := webclients.New(webclients.Config{
		UsersAddr:    cfg.usersAddr,
		ProductsAddr: cfg.productsAddr,
		OrdersAddr:   cfg.ordersAddr,
		CartAddr:     cfg.cartAddr,
		PaymentAddr:  cfg.paymentAddr,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer deps.Close()

	h := handlers.New(
		deps.Users,
		deps.Products,
		deps.Orders,
		deps.Cart,
		deps.Payment,
		authconfig.DefaultConfig(cfg.jwtSecret),
	)

	srv := &http.Server{
		Addr:    cfg.addr,
		Handler: newRouter(h),
	}
	runServer(srv)
}
