FROM golang:1.26.1-alpine AS builder

WORKDIR /src

COPY shared ./shared
COPY services/cart ./services/cart

WORKDIR /src/services/cart
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/cart ./cmd/cart

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/cart /app/cart

EXPOSE 9094

ENTRYPOINT ["/app/cart"]
