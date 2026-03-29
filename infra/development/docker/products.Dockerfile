FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY services ./services

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/products ./services/products/cmd/products

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/products /app/products

EXPOSE 8082

ENTRYPOINT ["/app/products"]
