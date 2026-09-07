package runtime

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type GRPCServerConfig struct {
	Addr            string
	ServiceName     string
	Register        func(*grpc.Server)
	ShutdownTimeout time.Duration
}

func ServeGRPC(ctx context.Context, cfg GRPCServerConfig) error {
	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}

	opts := append(sharedtrace.GRPCServerOptions(), grpc.ChainUnaryInterceptor(unaryAccessLog))
	server := grpc.NewServer(opts...)
	cfg.Register(server)
	reflection.Register(server)

	shutdownTimeout := cfg.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 30 * time.Second
	}
	errCh := make(chan error, 1)
	go func() { errCh <- server.Serve(lis) }()
	sharedlog.Info("starting grpc service", "addr", cfg.Addr)
	select {
	case err := <-errCh:
		server.Stop()
		return err
	case <-ctx.Done():
	}
	stopped := make(chan struct{})
	go func() { server.GracefulStop(); close(stopped) }()
	timer := time.NewTimer(shutdownTimeout)
	defer timer.Stop()
	select {
	case <-stopped:
	case <-timer.C:
		server.Stop()
		<-stopped
	}
	err = <-errCh
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}

func unaryAccessLog(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	st, _ := status.FromError(err)
	method := info.FullMethod
	if i := strings.LastIndex(method, "/"); i >= 0 && i+1 < len(method) {
		method = method[i+1:]
	}
	msg := fmt.Sprintf("%s %s", method, st.Code().String())
	args := []any{
		"grpc_method", info.FullMethod,
		"grpc_code", st.Code().String(),
		"duration_ms", time.Since(start).Milliseconds(),
	}
	// Access logs stay structured; non-OK codes use Warn so Explore level
	// filters surface failures without treating every RPC as an Error.
	if st.Code() != codes.OK {
		sharedlog.WarnContext(ctx, msg, args...)
	} else {
		sharedlog.InfoContext(ctx, msg, args...)
	}
	return resp, err
}
