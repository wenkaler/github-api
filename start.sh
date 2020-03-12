#!/bin/sh
docker stop gapi
docker rm gapi
docker run \
  --restart=always \
  --name gapi \
  -e GIT_TOKEN="607813a3e18ef52794e988c77c683eed71950a4d" \
  -e GIT_PROJECT="github-api" \
  -e GIT_OWNER="wenkaer" \
  -e GIT_EMAIL="kola1403@yandex.ru" \
  -e GIT_NAME="Chernyy Dmitriy" \
  -e TIME="14:29" \
  -d gapi:latest
sleep 0.1
docker logs gapi -f