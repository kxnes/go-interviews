#!/usr/bin/env bash

if [[ ! -f "go.mod" ]]; then
  echo "Usage from repo root: ./scripts/checker.sh"
  exit 1
fi

golangci-lint run ./...
golint ./...

if [[ $1 == "-t" ]]; then
    go test -cover ./...
fi
