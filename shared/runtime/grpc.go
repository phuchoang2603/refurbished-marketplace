package runtime

import (
	"context"
	"log"
	"net"

	sharedtrace "refurbished-marketplace/shared/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	opts := sharedtrace.GRPCServerOptions()
	server := grpc.NewServer(opts...)
	cfg.Register(server)
	reflection.Register(server)

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	log.Printf("starting %s grpc service on %s", cfg.ServiceName, cfg.Addr)
	return server.Serve(lis)
}
