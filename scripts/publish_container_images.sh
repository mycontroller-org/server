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

PLATFORMS=(
  "linux/amd64:linux-amd64"
  "linux/arm64:linux-arm64"
  "linux/arm/v7:linux-armv7"
  "linux/arm/v6:linux-armv6"
)

for spec in "${PLATFORMS[@]}"; do
  platform="${spec%%:*}"
  binary_dir="${spec##*:}"
  binary_path="builds/binary/${binary_dir}/mycontroller-${TARGET_BINARY}"
  if [ ! -f "${binary_path}" ]; then
    echo "missing ${binary_path}; run ./scripts/generate_executables.sh first" >&2
    exit 1
  fi

  tags=()
  for registry in "${REGISTRIES[@]}"; do
    tags+=(--tag "${registry}/${TARGET_BINARY}:${IMAGE_TAG}-${binary_dir}")
  done

  docker buildx build --push \
    --progress=plain \
    --provenance=false \
    --sbom=false \
    --platform "${platform}" \
    --file "docker/${TARGET_BINARY}.Dockerfile" \
    --build-arg "BINARY_PATH=${binary_path}" \
    "${tags[@]}" \
    .
done

for registry in "${REGISTRIES[@]}"; do
  docker buildx imagetools create \
    --tag "${registry}/${TARGET_BINARY}:${IMAGE_TAG}" \
    "${registry}/${TARGET_BINARY}:${IMAGE_TAG}-linux-amd64" \
    "${registry}/${TARGET_BINARY}:${IMAGE_TAG}-linux-arm64" \
    "${registry}/${TARGET_BINARY}:${IMAGE_TAG}-linux-armv7" \
    "${registry}/${TARGET_BINARY}:${IMAGE_TAG}-linux-armv6"
done
