module github.com/phuchoang2603/refurbished-marketplace/shared/messaging

go 1.26.5

require (
	github.com/phuchoang2603/refurbished-marketplace/shared/observe/log v0.0.0
	github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace v0.0.0
	github.com/twmb/franz-go v1.21.5
	go.opentelemetry.io/otel v1.46.0
	go.opentelemetry.io/otel/trace v1.46.0
	golang.org/x/sync v0.22.0
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0 // indirect
	github.com/klauspost/compress v1.19.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.28 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.13.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.70.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.45.0 // indirect
	go.opentelemetry.io/otel/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk v1.46.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.0 // indirect
	golang.org/x/crypto v0.55.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260810153831-ec0a7760b754 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260810153831-ec0a7760b754 // indirect
	google.golang.org/grpc v1.83.0 // indirect
)

replace (
	github.com/phuchoang2603/refurbished-marketplace/shared/auth => ../auth
	github.com/phuchoang2603/refurbished-marketplace/shared/err/dberr => ../err/dberr
	github.com/phuchoang2603/refurbished-marketplace/shared/err/grpcerr => ../err/grpcerr
	github.com/phuchoang2603/refurbished-marketplace/shared/observe/log => ../observe/log
	github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace => ../observe/trace
	github.com/phuchoang2603/refurbished-marketplace/shared/proto => ../proto
	github.com/phuchoang2603/refurbished-marketplace/shared/runtime => ../runtime
	github.com/phuchoang2603/refurbished-marketplace/shared/testutil/kafka => ../testutil/kafka
	github.com/phuchoang2603/refurbished-marketplace/shared/testutil/postgres => ../testutil/postgres
	github.com/phuchoang2603/refurbished-marketplace/shared/testutil/redis => ../testutil/redis
)
