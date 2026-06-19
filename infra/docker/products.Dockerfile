# Wrapper around go-service.Dockerfile with products defaults.
# Build context must be the repository root.

FROM golang:1.26.2-alpine AS builder

ARG BUILD_PKG=./services/products/cmd/products
ARG BUILD_BIN=products

WORKDIR /src

COPY go.work go.work.sum ./
COPY shared ./shared
COPY services ./services
COPY tools ./tools

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/${BUILD_BIN} ${BUILD_PKG}

FROM gcr.io/distroless/static-debian12

ARG BUILD_BIN=products

WORKDIR /app

COPY --from=builder /out/${BUILD_BIN} /app/service

EXPOSE 9092

ENTRYPOINT ["/app/service"]
