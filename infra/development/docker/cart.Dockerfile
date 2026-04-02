FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY services ./services
COPY shared ./shared

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/cart ./services/cart/cmd/cart

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/cart /app/cart

EXPOSE 9094

ENTRYPOINT ["/app/cart"]
