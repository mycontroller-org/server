#!/bin/bash

source ./scripts/version.sh

# container registry
REGISTRY='quay.io/mycontroller'
ALT_REGISTRY='docker.io/mycontroller'
PLATFORMS="linux/arm/v6,linux/arm/v7,linux/arm64,linux/amd64"
IMAGE_TAG=${VERSION}

# debug lines
echo $PWD
ls -alh
git branch

TARGET_BINARY=${TARGET_BUILD:-server}

# build and push to quay.io
docker buildx build --push \
  --progress=plain \
  --build-arg=GOPROXY=${GOPROXY} \
  --platform ${PLATFORMS} \
  --file docker/${TARGET_BINARY}.Dockerfile \
  --tag ${REGISTRY}/${TARGET_BINARY}:${IMAGE_TAG} .

# build and push to docker.io
docker buildx build --push \
  --progress=plain \
  --build-arg=GOPROXY=${GOPROXY} \
  --platform ${PLATFORMS} \
  --file docker/${TARGET_BINARY}.Dockerfile \
  --tag ${ALT_REGISTRY}/${TARGET_BINARY}:${IMAGE_TAG} .

