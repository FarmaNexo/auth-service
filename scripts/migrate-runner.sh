#!/bin/sh
# URL-encode-aware migration runner for golang-migrate.
# Reads env vars (inyectadas por ECS via Secrets Manager):
#   DB_HOST, DB_USER, DB_PASSWORD, DB_NAME
# Optional:
#   DB_PORT (default 5432), DB_SSLMODE (default require),
#   MIGRATIONS_PATH (default /app/migrations)

set -eu

: "${DB_HOST:?missing DB_HOST}"
: "${DB_USER:?missing DB_USER}"
: "${DB_PASSWORD:?missing DB_PASSWORD}"
: "${DB_NAME:?missing DB_NAME}"

MIGRATIONS_PATH="${MIGRATIONS_PATH:-/app/migrations}"
DB_PORT="${DB_PORT:-5432}"
DB_SSLMODE="${DB_SSLMODE:-require}"

# POSIX sh URL-encoder — unreserved chars [A-Za-z0-9._~-] pasan, el resto se
# codifica como %HH. Funciona en busybox ash (alpine).
urlenc() {
  s="$1"
  i=1
  len=${#s}
  while [ "$i" -le "$len" ]; do
    c=$(printf '%s' "$s" | cut -c "$i")
    case "$c" in
      [a-zA-Z0-9._~-]) printf '%s' "$c" ;;
      *) printf '%%%02X' "'$c" ;;
    esac
    i=$((i + 1))
  done
}

ENC_USER=$(urlenc "$DB_USER")
ENC_PASS=$(urlenc "$DB_PASSWORD")

DATABASE_URL="postgres://${ENC_USER}:${ENC_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

echo "Running migrate up against ${DB_HOST}:${DB_PORT}/${DB_NAME} (sslmode=${DB_SSLMODE})"
echo "Migrations path: ${MIGRATIONS_PATH}"

exec migrate -path "${MIGRATIONS_PATH}" -database "${DATABASE_URL}" up
