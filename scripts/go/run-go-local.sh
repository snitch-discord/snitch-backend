#!/bin/bash

BASE_DIR=$(dirname "$0")/../..
PRIVATE_KEY=$(<"${BASE_DIR}/scripts/secrets/snitch_private_key.pem")

export LIBSQL_HOST=localhost
export LIBSQL_PORT=8080
export LIBSQL_ADMIN_PORT=9090
export LIBSQL_AUTH_KEY="$PRIVATE_KEY"

go run "${BASE_DIR}"/cmd/snitchbe/main.go
