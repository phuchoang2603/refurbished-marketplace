FROM golang:1.26.1-alpine AS builder

WORKDIR /src

COPY shared ./shared
COPY services/payment ./services/payment

WORKDIR /src/services/payment
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/payment ./cmd/payment

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/payment /app/payment

EXPOSE 9096

ENTRYPOINT ["/app/payment"]
