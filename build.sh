#!/bin/sh
docker rm -f translate
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o translate translate
docker build -t translate .
docker run -d -name translate -P translate
