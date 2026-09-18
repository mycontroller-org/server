#!/bin/bash

# Push container images from host-built linux binaries in builds/binary/.
# Build those first with ./scripts/generate_executables.sh

set -euo pipefail

source ./scripts/version.sh

REGISTRY='quay.io/mycontroller'
ALT_REGISTRY='docker.io/mycontroller'
GHCR_REGISTRY="ghcr.io/${GITHUB_REPOSITORY_OWNER:-mycontroller-org}"
IMAGE_TAG=${VERSION}
TARGET_BINARY=${TARGET_BUILD:-server}

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

  docker buildx build --push \
    --progress=plain \
    --platform "${platform}" \
    --file "docker/${TARGET_BINARY}.Dockerfile" \
    --build-arg "BINARY_PATH=${binary_path}" \
    --tag "${REGISTRY}/${TARGET_BINARY}:${IMAGE_TAG}-${binary_dir}" \
    --tag "${ALT_REGISTRY}/${TARGET_BINARY}:${IMAGE_TAG}-${binary_dir}" \
    --tag "${GHCR_REGISTRY}/${TARGET_BINARY}:${IMAGE_TAG}-${binary_dir}" \
    .
done

for registry in "${REGISTRY}" "${ALT_REGISTRY}" "${GHCR_REGISTRY}"; do
  docker buildx imagetools create \
    --tag "${registry}/${TARGET_BINARY}:${IMAGE_TAG}" \
    "${registry}/${TARGET_BINARY}:${IMAGE_TAG}-linux-amd64" \
    "${registry}/${TARGET_BINARY}:${IMAGE_TAG}-linux-arm64" \
    "${registry}/${TARGET_BINARY}:${IMAGE_TAG}-linux-armv7" \
    "${registry}/${TARGET_BINARY}:${IMAGE_TAG}-linux-armv6"
done
