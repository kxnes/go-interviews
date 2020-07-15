#!/usr/bin/env bash

go build -o ./dist/md5calc ./cmd/md5calc

docker-compose -f deployments/docker-compose.dev.yml --project-directory . up --build
