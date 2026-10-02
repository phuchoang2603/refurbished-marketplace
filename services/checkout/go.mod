module github.com/phuchoang2603/refurbished-marketplace/services/checkout

go 1.26.5

require (
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.12.3
	github.com/phuchoang2603/refurbished-marketplace/shared/messaging v0.0.0
	github.com/phuchoang2603/refurbished-marketplace/shared/observe/log v0.0.0
	github.com/phuchoang2603/refurbished-marketplace/shared/proto v0.0.0
	github.com/phuchoang2603/refurbished-marketplace/shared/runtime v0.0.0
	github.com/phuchoang2603/refurbished-marketplace/shared/testutil/postgres v0.0.0
	go.opentelemetry.io/otel v1.46.0
	go.opentelemetry.io/otel/metric v1.46.0
	google.golang.org/grpc v1.83.0
	google.golang.org/protobuf v1.36.12
)
