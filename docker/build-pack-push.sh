#! /bin/bash

set -uexo pipefail

echo "This is for local development use only"

REPO=tim-hm
TAG=sha-9e08649

go build -o build/ubuntu-reportd ./cmd/ubuntu-reportd
docker build -f docker/ubuntu-reportd/Dockerfile -t ghcr.io/$REPO/ubuntu-report:$TAG .
docker push ghcr.io/$REPO/ubuntu-report:$TAG
