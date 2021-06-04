#!/bin/bash

# container registry
REGISTRY='quay.io/mycontroller-org'
ALT_REGISTRY='dokcer.io/mycontroller'
IMAGE_ALL_IN_ONE="all-in-one"
IMAGE_CORE="core"
IMAGE_GATEWAY="gateway"
PLATFORMS="linux/arm/v6,linux/arm/v7,linux/arm64,linux/amd64"
#IMAGE_TAG="master"  # application tag
IMAGE_TAG=`git rev-parse --abbrev-ref HEAD`

# debug lines
echo $PWD
ls -alh
git branch

TARGET_BINARY=${TARGET_BUILD:-all-in-one}

# build conatiner image
docker buildx build --push \
  --progress=plain \
  --build-arg=GOPROXY=${GOPROXY} \
  --platform ${PLATFORMS} \
  --file docker/${TARGET_BINARY}.Dockerfile \
  --tag ${REGISTRY}/${IMAGE_CORE}:${IMAGE_TAG} .

# copy image into docker hub
skopeo copy docker://${REGISTRY}/${TARGET_BINARY}:${IMAGE_TAG} docker://${ALT_REGISTRY}/${TARGET_BINARY}:${IMAGE_TAG}
