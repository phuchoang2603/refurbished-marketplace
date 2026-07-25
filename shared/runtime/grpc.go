package runtime

import (
	"context"
	"log/slog"
	"net"
	"time"

	sharedtrace "refurbished-marketplace/shared/observe/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type GRPCServerConfig struct {
	Addr        string
	ServiceName string
	Register    func(*grpc.Server)
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

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	slog.Info("starting grpc service", "addr", cfg.Addr)
	return server.Serve(lis)
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
	slog.LogAttrs(
		ctx, slog.LevelInfo, "grpc request",
		slog.String("grpc_method", info.FullMethod),
		slog.String("grpc_code", st.Code().String()),
		slog.Int64("duration_ms", time.Since(start).Milliseconds()),
	)
	return resp, err
}
