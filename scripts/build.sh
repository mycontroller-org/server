#!/bin/bash

# container registry
REGISTRY='quay.io/mycontroller-org'
IMAGE_ALL_IN_ONE="${REGISTRY}/all-in-one"
IMAGE_CORE="${REGISTRY}/core"
IMAGE_GATEWAY="${REGISTRY}/gateway"
#IMAGE_TAG="master"  # application tag
IMAGE_TAG=`git rev-parse --abbrev-ref HEAD`

# debug lines
echo $PWD
ls -alh
git branch

TARGET_BUILD=${TARGET_BUILD:-all-in-one}

# build conatiner images
if [[ "$TARGET_BINARY" == "core" ]]; then
  # build core image
  docker buildx build --push --progress=plain --build-arg=GOPROXY=${GOPROXY} --platform linux/arm/v6,linux/arm/v7,linux/arm64,linux/amd64 --file docker/core.Dockerfile --tag ${IMAGE_CORE}:${IMAGE_TAG} .

elif [[ "$TARGET_BINARY" == "gateway" ]]; then
  # build gateway image
  docker buildx build --push --progress=plain --build-arg=GOPROXY=${GOPROXY} --platform linux/arm/v6,linux/arm/v7,linux/arm64,linux/amd64 --file docker/gateway.Dockerfile --tag ${IMAGE_ALL_IN_ONE}:${IMAGE_TAG} .

else
  # build all-in-one image
  docker buildx build --push --progress=plain --build-arg=GOPROXY=${GOPROXY} --platform linux/arm/v6,linux/arm/v7,linux/arm64,linux/amd64 --file docker/all-in-one.Dockerfile --tag ${IMAGE_GATEWAY}:${IMAGE_TAG} .
fi