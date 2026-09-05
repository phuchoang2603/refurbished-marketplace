package clients

import (
	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newConn(addr string) (*grpc.ClientConn, error) {
	opts := append([]grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, sharedtrace.GRPCDialOptions()...)
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}
	// grpc.NewClient is lazy; Connect now so checkout is not the first
	// RPC (Cilium mTLS + HTTP/2 setup) to orders/payment.
	conn.Connect()
	return conn, nil
}
