#! /bin/bash

set -uexo pipefail

echo "This is for local development use only"

REPO=tim-hm
SHORT_SHA=$(git rev-parse --short HEAD)
TAG="sha-${SHORT_SHA}"

go build -o build/ubuntu-reportd ./cmd/ubuntu-reportd
docker build -f docker/ubuntu-reportd/Dockerfile -t "ghcr.io/${REPO}/ubuntu-report:${TAG}" .
docker push "ghcr.io/${REPO}/ubuntu-report:${TAG}"
