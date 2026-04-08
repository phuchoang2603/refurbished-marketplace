FROM golang:1.26.1-alpine AS builder

WORKDIR /src

COPY shared ./shared
COPY services/inventory ./services/inventory

WORKDIR /src/services/inventory
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/inventory ./cmd/inventory

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/inventory /app/inventory

EXPOSE 8083

ENTRYPOINT ["/app/inventory"]
