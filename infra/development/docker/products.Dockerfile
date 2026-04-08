FROM golang:1.26.1-alpine AS builder

WORKDIR /src

COPY shared ./shared
COPY services/products ./services/products

WORKDIR /src/services/products
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/products ./cmd/products

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/products /app/products

EXPOSE 8082

ENTRYPOINT ["/app/products"]
