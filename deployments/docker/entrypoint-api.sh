#!/bin/sh
set -eu

: "${MYSQL_PASSWORD:?MYSQL_PASSWORD is required}"
: "${JWT_SECRET:?JWT_SECRET is required}"

export MYSQL_PASSWORD
export JWT_SECRET
export REDIS_PASSWORD="${REDIS_PASSWORD:-}"
export CORS_ORIGIN="${CORS_ORIGIN:-http://localhost}"

envsubst '${MYSQL_PASSWORD} ${JWT_SECRET} ${REDIS_PASSWORD} ${CORS_ORIGIN}' \
  < /app/configs/config.prod.docker.yaml.tmpl \
  > /app/configs/config.prod.yaml

exec ./arlo-admin --config=configs/config.yaml
