package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	webclients "refurbished-marketplace/services/web/internal/clients"
	"refurbished-marketplace/services/web/internal/config"
	"refurbished-marketplace/services/web/internal/handlers"
	sharedhandlers "refurbished-marketplace/services/web/internal/handlers/shared"
	authconfig "refurbished-marketplace/shared/auth/config"
	sharedlog "refurbished-marketplace/shared/observe/log"
	"refurbished-marketplace/shared/runtime"
)

func main() {
	runtime.InitLogging("web")
	cfg := config.LoadConfig()
	if err := config.ValidateConfig(cfg); err != nil {
		sharedlog.Fatal("config", "err", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := runtime.InitTracing(ctx, "web")
	if err != nil {
		sharedlog.Fatal("init tracing", "err", err)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			sharedlog.Error("tracing shutdown", "err", err)
		}
	}()

	deps, err := webclients.New(webclients.Config{
		UsersAddr:    cfg.UsersAddr,
		ProductsAddr: cfg.ProductsAddr,
		OrdersAddr:   cfg.OrdersAddr,
		CartAddr:     cfg.CartAddr,
		PaymentAddr:  cfg.PaymentAddr,
	})
	if err != nil {
		sharedlog.Fatal("clients", "err", err)
	}
	defer deps.Close()

	h := handlers.New(
		deps.Users,
		deps.Products,
		deps.Orders,
		deps.Cart,
		deps.Payment,
		sharedhandlers.HostedPaymentConfig{
			GatewayBaseURL:  cfg.GatewayBaseURL,
			PublicBaseURL:   cfg.PublicBaseURL,
			CallbackBaseURL: cfg.CallbackBaseURL,
		},
		authconfig.DefaultConfig(cfg.JWTSecret),
	)

	if err := runtime.ServeHTTP(ctx, runtime.HTTPServerConfig{
		Addr:        cfg.HTTPAddr,
		ServiceName: "web",
		Handler:     newRouter(h),
	}); err != nil {
		sharedlog.Fatal("http serve", "err", err)
	}
}
