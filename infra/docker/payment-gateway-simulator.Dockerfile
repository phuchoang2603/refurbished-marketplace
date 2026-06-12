FROM golang:1.26.2-alpine AS builder

WORKDIR /src

COPY tools/payment-gateway-simulator ./tools/payment-gateway-simulator

WORKDIR /src/tools/payment-gateway-simulator
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/payment-gateway-simulator .

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/payment-gateway-simulator /app/payment-gateway-simulator

EXPOSE 8097

ENTRYPOINT ["/app/payment-gateway-simulator"]
