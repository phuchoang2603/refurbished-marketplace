# Generic Go service image built from the repo go.work workspace.
# Build context must be the repository root.
#
# Example:
#   docker build -f infra/docker/go-service.Dockerfile \
#     --build-arg BUILD_PKG=./services/users/cmd/users \
#     --build-arg BUILD_BIN=users \
#     --build-arg EXPOSE_PORT=9091 \
#     -t users .
#
# Service build args (BUILD_PKG / BUILD_BIN / EXPOSE_PORT):
#   users                     ./services/users/cmd/users                     users                     9091
#   products                  ./services/products/cmd/products               products                  9092
#   orders                    ./services/orders/cmd/orders                   orders                    9093
#   cart                      ./services/cart/cmd/cart                       cart                      9094
#   payment                   ./services/payment/cmd/payment                 payment                   9096
#   payment-gateway-simulator ./tools/payment-gateway-simulator              payment-gateway-simulator 8097

FROM golang:1.26.2-alpine AS builder

ARG BUILD_PKG
ARG BUILD_BIN

WORKDIR /src

COPY go.work go.work.sum ./
COPY shared ./shared
COPY services ./services
COPY tools ./tools

RUN go mod download

RUN test -n "$BUILD_PKG" && test -n "$BUILD_BIN"
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/${BUILD_BIN} ${BUILD_PKG}

FROM gcr.io/distroless/static-debian12

ARG BUILD_BIN
ARG EXPOSE_PORT

WORKDIR /app

COPY --from=builder /out/${BUILD_BIN} /app/service

EXPOSE ${EXPOSE_PORT}

ENTRYPOINT ["/app/service"]
