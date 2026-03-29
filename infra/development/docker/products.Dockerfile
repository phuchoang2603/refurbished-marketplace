FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY build/products /app/products

EXPOSE 8082

ENTRYPOINT ["/app/products"]
