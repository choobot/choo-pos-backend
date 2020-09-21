#!/bin/sh

source .env

docker-compose exec go "/bin/sh" "-c" "go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out"