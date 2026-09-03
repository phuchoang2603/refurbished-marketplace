package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	webclients "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/clients"
	"github.com/phuchoang2603/refurbished-marketplace/services/web/internal/config"
	"github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers"
	sharedhandlers "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers/shared"
	authconfig "github.com/phuchoang2603/refurbished-marketplace/shared/auth/config"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	"github.com/phuchoang2603/refurbished-marketplace/shared/runtime"
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

	shutdownMetrics, err := runtime.InitMetrics(ctx, "web")
	if err != nil {
		sharedlog.Fatal("init metrics", "err", err)
	}
	defer func() {
		if err := shutdownMetrics(context.Background()); err != nil {
			sharedlog.Error("metrics shutdown", "err", err)
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
