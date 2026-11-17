FROM golang:1.26.5-alpine AS builder

ARG BUILD_PKG
ARG BUILD_BIN

WORKDIR /src

COPY go.work go.work.sum ./
COPY shared ./shared
COPY services ./services

RUN go mod download

RUN test -n "$BUILD_PKG" && test -n "$BUILD_BIN"
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/${BUILD_BIN} ${BUILD_PKG}

FROM gcr.io/distroless/static-debian12

ARG BUILD_BIN
ARG EXPOSE_PORT

WORKDIR /app

COPY --from=builder /out/${BUILD_BIN} /app/service

EXPOSE ${EXPOSE_PORT}

ENTRYPOINT ["/app/service"]
