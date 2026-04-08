FROM golang:1.26.1-alpine AS builder

WORKDIR /src

COPY shared ./shared
COPY services/users ./services/users

WORKDIR /src/services/users
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/users ./cmd/users

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/users /app/users

EXPOSE 9091

ENTRYPOINT ["/app/users"]
