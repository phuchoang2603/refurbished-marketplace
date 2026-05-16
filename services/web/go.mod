module refurbished-marketplace/services/web

go 1.26.1

require (
	github.com/a-h/templ v0.3.1001
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/starfederation/datastar-go v1.2.1
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
	refurbished-marketplace/shared/auth v0.0.0
	refurbished-marketplace/shared/proto v0.0.0
)

replace refurbished-marketplace/shared/auth => ../../shared/auth

replace refurbished-marketplace/shared/proto => ../../shared/proto

require (
	github.com/CAFxX/httpcompression v0.0.9 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260217215200-42d3e9bedb6d // indirect
)
