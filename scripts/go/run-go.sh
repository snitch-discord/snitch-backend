#!/bin/bash

BASE_DIR=$(dirname "$0")/../..
PRIVATE_KEY=$(<"${BASE_DIR}/scripts/secrets/snitch_private_key.pem")
GO_IMAGE_NAME=snitchbe

docker stop ${GO_IMAGE_NAME}
docker container rm ${GO_IMAGE_NAME}

docker build -t ${GO_IMAGE_NAME} -f "${BASE_DIR}"/Containerfile .

docker run -d --name ${GO_IMAGE_NAME} \
  --network snitch-network \
  -p 8080:8080 \
  -e LIBSQL_HOST=snitch-sqld \
  -e LIBSQL_PORT=8081 \
  -e LIBSQL_AUTH_KEY="$PRIVATE_KEY" \
  ${GO_IMAGE_NAME}
