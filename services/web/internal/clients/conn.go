package clients

import (
	sharedtrace "refurbished-marketplace/shared/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newConn(addr string) (*grpc.ClientConn, error) {
	opts := append([]grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, sharedtrace.GRPCDialOptions()...)
	return grpc.NewClient(addr, opts...)
}
