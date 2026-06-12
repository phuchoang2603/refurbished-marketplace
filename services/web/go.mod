module refurbished-marketplace/services/web

go 1.26.2

require (
	github.com/Oudwins/tailwind-merge-go v0.2.0
	github.com/a-h/templ v0.3.1001
	github.com/go-chi/chi/v5 v5.2.3
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/starfederation/datastar-go v1.2.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.65.0
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
	refurbished-marketplace/shared/auth v0.0.0
	refurbished-marketplace/shared/proto v0.0.0
)

require (
	github.com/CAFxX/httpcompression v0.0.9 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.25 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.41.0 // indirect
	go.opentelemetry.io/otel/metric v1.41.0 // indirect
	go.opentelemetry.io/otel/sdk v1.41.0 // indirect
	go.opentelemetry.io/otel/trace v1.41.0 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
)

replace refurbished-marketplace/shared/auth => ../../shared/auth

replace refurbished-marketplace/shared/proto => ../../shared/proto
