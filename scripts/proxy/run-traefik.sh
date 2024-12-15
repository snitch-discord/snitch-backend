#!/bin/bash

BASE_DIR=$(dirname "$0")/../..

PROXY_DIR="${BASE_DIR}/scripts/proxy"
PROXY_IMAGE_NAME=snitch-traefik

docker stop ${PROXY_IMAGE_NAME}
docker container rm ${PROXY_IMAGE_NAME}

docker run -p 8080:8080 -p 9090:9090 -d --name ${PROXY_IMAGE_NAME} -ti \
  --network snitch-network \
  -v "${PROXY_DIR}/traefik.yml":/etc/traefik/traefik.yml \
  traefik:latest
