FROM golang:1.26.1-alpine AS builder

WORKDIR /src

COPY shared ./shared
COPY services/web ./services/web

WORKDIR /src/services/web
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/web ./cmd/web

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/web /app/web

EXPOSE 8080

ENTRYPOINT ["/app/web"]
