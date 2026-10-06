package trace

import (
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// GRPCServerOptions returns server options that install OpenTelemetry stats handling
// and accept the idle keepalive pings sent by GRPCDialOptions.
func GRPCServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		// MinTime must stay below the client keepalive Time, otherwise the server
		// replies GOAWAY ENHANCE_YOUR_CALM "too_many_pings".
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             15 * time.Second,
			PermitWithoutStream: true,
		}),
	}
}

// GRPCDialOptions returns dial options that install OpenTelemetry stats handling
// and HTTP/2 keepalives so mesh mTLS handshakes are not paid on every idle reconnect.
func GRPCDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	}
}
