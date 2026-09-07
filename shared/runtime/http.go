package runtime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
)

const defaultHTTPShutdownTimeout = 30 * time.Second

type HTTPServerConfig struct {
	Addr            string
	ServiceName     string
	Handler         http.Handler
	ShutdownTimeout time.Duration
}

func ServeHTTP(ctx context.Context, cfg HTTPServerConfig) error {
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           cfg.Handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownTimeout := cfg.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = defaultHTTPShutdownTimeout
	}

	errCh := make(chan error, 1)
	go func() {
		sharedlog.Info("starting http service", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return errors.Join(fmt.Errorf("http shutdown: %w", err), srv.Close())
	}
	return nil
}
