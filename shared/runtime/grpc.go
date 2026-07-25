package runtime

import (
	"context"
	"net"
	"time"

	sharedlog "refurbished-marketplace/shared/observe/log"
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

	sharedlog.Info("starting grpc service", "addr", cfg.Addr)
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
	sharedlog.InfoContext(
		ctx, "grpc request",
		"grpc_method", info.FullMethod,
		"grpc_code", st.Code().String(),
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return resp, err
}
