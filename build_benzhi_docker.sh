#!/usr/bin/env bash
set -euo pipefail
NAME="${1:?image name required}"
PLATFORM="${2:-linux/amd64}"
IMAGE="benzhi/${NAME}:latest"
docker build --platform "${PLATFORM}" -f benzhi.Dockerfile -t "${IMAGE}" .
