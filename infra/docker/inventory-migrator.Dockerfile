FROM ghcr.io/kukymbr/goose-docker:3.27.0

COPY services/inventory/db/migrations /migrations
