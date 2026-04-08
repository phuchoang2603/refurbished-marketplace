module refurbished-marketplace/services/web

go 1.26.1

require (
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
	refurbished-marketplace/shared v0.0.0
)

replace refurbished-marketplace/shared => ../../shared

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260217215200-42d3e9bedb6d // indirect
)
