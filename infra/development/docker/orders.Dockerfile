FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY build/orders /app/orders

EXPOSE 8083

ENTRYPOINT ["/app/orders"]
