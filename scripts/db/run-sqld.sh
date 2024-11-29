#!/bin/bash

BASE_DIR=$(dirname "$0")/../..

PUBLIC_KEY=$(<"${BASE_DIR}/scripts/secrets/snitch_public_key.pem")

DB_DIR="${BASE_DIR}/scripts/db"
DB_IMAGE_NAME=snitch-sqld
DB_FOLDER_NAME=snitch-sqld-data

DB_DATA_PATH="${DB_DIR}/${DB_FOLDER_NAME}"

docker stop ${DB_IMAGE_NAME}
docker container rm ${DB_IMAGE_NAME}
sudo rm -rf "${DB_DATA_PATH}"

mkdir "${DB_DATA_PATH}"

docker run -p 8081:8080 -p 9091:9090 -d --name ${DB_IMAGE_NAME} -ti \
  --network snitch-network \
  -e SQLD_NODE=primary \
  -e SQLD_DB_PATH=snitch.db \
  -e SQLD_AUTH_JWT_KEY="${PUBLIC_KEY}" \
  -e RUST_LOG=trace \
  -v "${DB_DATA_PATH}":/var/lib/sqld \
  ghcr.io/tursodatabase/libsql-server:latest \
  sqld --admin-listen-addr 0.0.0.0:9090 --enable-namespaces
