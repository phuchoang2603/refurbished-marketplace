# Generic goose migration image. Build context must be the repository root.
#
# Example:
#   docker build -f infra/docker/goose-migrator.Dockerfile \
#     --build-arg MIGRATIONS_DIR=services/users/db/migrations \
#     -t users-migrator .

FROM ghcr.io/kukymbr/goose-docker:3.27.0

ARG MIGRATIONS_DIR
RUN test -n "$MIGRATIONS_DIR"

COPY ${MIGRATIONS_DIR} /migrations
