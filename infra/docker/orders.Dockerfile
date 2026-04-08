FROM golang:1.26.1-alpine AS builder

WORKDIR /src

COPY shared ./shared
COPY services/orders ./services/orders

WORKDIR /src/services/orders
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/orders ./cmd/orders

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/orders /app/orders

EXPOSE 8083

ENTRYPOINT ["/app/orders"]
