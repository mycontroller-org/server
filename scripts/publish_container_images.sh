#!/bin/bash

# Push container images from host-built linux binaries in builds/binary/.
# Build those first with ./scripts/generate_executables.sh

set -euo pipefail

source ./scripts/version.sh

GHCR_REGISTRY="ghcr.io/${GITHUB_REPOSITORY_OWNER:-mycontroller-org}"
IMAGE_TAG=${VERSION}
TARGET_BINARY=${TARGET_BUILD:-server}

REGISTRIES=("${GHCR_REGISTRY}")
if [ "${PUSH_DOCKER_QUAY:-0}" = "1" ]; then
  REGISTRIES+=("quay.io/mycontroller" "docker.io/mycontroller")
fi

for dir in linux-amd64 linux-arm64 linux-armv7 linux-armv6; do
  binary_path="builds/binary/${dir}/mycontroller-${TARGET_BINARY}"
  if [ ! -f "${binary_path}" ]; then
    echo "missing ${binary_path}; run ./scripts/generate_executables.sh first" >&2
    exit 1
  fi
done

tags=()
for registry in "${REGISTRIES[@]}"; do
  tags+=(--tag "${registry}/${TARGET_BINARY}:${IMAGE_TAG}")
done

docker buildx build --push \
  --progress=plain \
  --provenance=false \
  --sbom=false \
  --platform linux/amd64,linux/arm64,linux/arm/v7,linux/arm/v6 \
  --file "docker/${TARGET_BINARY}.Dockerfile" \
  "${tags[@]}" \
  .
