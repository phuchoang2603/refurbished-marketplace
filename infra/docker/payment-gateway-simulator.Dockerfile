# Payment gateway simulator — workspace build.
# Build context must be the repository root.

FROM golang:1.26.2-alpine AS builder

ARG BUILD_PKG=./tools/payment-gateway-simulator
ARG BUILD_BIN=payment-gateway-simulator

WORKDIR /src

COPY go.work go.work.sum ./
COPY shared ./shared
COPY services ./services
COPY tools ./tools

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/${BUILD_BIN} ${BUILD_PKG}

FROM gcr.io/distroless/static-debian12

ARG BUILD_BIN=payment-gateway-simulator

WORKDIR /app

COPY --from=builder /out/${BUILD_BIN} /app/service

EXPOSE 8097

ENTRYPOINT ["/app/service"]
