#!/bin/bash

BASE_DIR=$(dirname "$0")/../..
PROXY_IMAGE_NAME=snitch-traefik

docker stop ${PROXY_IMAGE_NAME}
docker container rm ${PROXY_IMAGE_NAME}

docker run -p 8080:8080 -p 9090:9090 -d --name ${PROXY_IMAGE_NAME} -ti \
  --network snitch-network \
  --label "traefik.http.services.snitch-sqld.loadbalancer.servers[0].url=http://snitch-sqld" \
  --label "traefik.http.routers.map-subdomain.rule=HostRegexp(\`{subdomain:[a-z0-9-.]+}\\.snitch-traefik\`)" \
  --label "traefik.http.routers.map-subdomain.service=snitch-sqld" \
  --label "traefik.http.routers.map-subdomain.entrypoints=web-8080" \
  --label "traefik.http.routers.map-subdomain.middlewares=redirect-sqld,add-host-header" \
  --label "traefik.http.routers.catch-all.rule=!HostRegexp(\`{subdomain:[a-z0-9-]+}\\.snitch-traefik\`)" \
  --label "traefik.http.routers.catch-all.service=snitch-sqld" \
  --label "traefik.http.routers.catch-all.entrypoints=web-8080,web-9090" \
  --label "traefik.http.middlewares.add-host-header.headers.customrequestheaders.Host={subdomain}.db" \
  --label "traefik.http.middlewares.redirect-sqld.redirectregex.regex=^http://([a-z0-9-]+)\\.snitch-traefik(.*)" \
  --label "traefik.http.middlewares.redirect-sqld.redirectregex.replacement=http://snitch-sqld\${2}" \
  --label "traefik.http.middlewares.redirect-sqld.redirectregex.permanent=true" \
  traefik:latest \
  --log.level=TRACE \
  --accesslog=true \
  --accesslog.addinternals \
  --entrypoints.web-8080.address=:8080 \
  --entrypoints.web-9090.address=:9090
