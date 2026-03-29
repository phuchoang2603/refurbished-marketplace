FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY services ./services

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/users ./services/users/cmd/users

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/users /app/users

EXPOSE 8081

ENTRYPOINT ["/app/users"]
