FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY build/users /app/users

EXPOSE 8081

ENTRYPOINT ["/app/users"]
