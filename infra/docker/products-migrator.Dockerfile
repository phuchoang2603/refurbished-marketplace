FROM ghcr.io/kukymbr/goose-docker:3.27.0

COPY services/products/db/migrations /migrations
