#!/bin/bash

BASE_DIR=$(dirname "$0")/..

docker network inspect snitch-network >/dev/null 2>&1 || \
  docker network create snitch-network

SECRETS_DIR="${BASE_DIR}/scripts/secrets"
PRIVATE_KEY_PATH="${SECRETS_DIR}/snitch_private_key.pem"
PUBLIC_KEY_PATH="${SECRETS_DIR}/snitch_public_key.pem"

if [ ! -d "$SECRETS_DIR" ]; then
  mkdir "$SECRETS_DIR"
  openssl genpkey -algorithm ed25519 -outform PEM -out "$PRIVATE_KEY_PATH"
  openssl pkey -in "$PRIVATE_KEY_PATH" -pubout > "$PUBLIC_KEY_PATH"
fi

bash "${BASE_DIR}"/scripts/db/run-sqld.sh
bash "${BASE_DIR}"/scripts/proxy/run-traefik.sh
bash "${BASE_DIR}"/scripts/go/run-go.sh
