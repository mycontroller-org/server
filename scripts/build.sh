#!/bin/bash

# container registry
REGISTRY='quay.io/mycontroller-org'
IMAGE_ALL_IN_ONE="${REGISTRY}/all-in-one"
IMAGE_CORE="${REGISTRY}/core"
IMAGE_GATEWAY="${REGISTRY}/gateway"
#IMAGE_TAG="master"  # application tag
IMAGE_TAG=`git rev-parse --abbrev-ref HEAD`

# alpine golang builder image
# GOLANG_BUILDER_IMAGE="quay.io/mycontroller-org/golang"
# GOLANG_BUILDER_TAG="1.16.0-alpine3.13"

# debug lines
echo $PWD
ls -alh
git branch

# build application inside continer
# docker run --rm \
#     -v "$PWD"/:/usr/src/mycontroller -w /usr/src/mycontroller \
#     ${GOLANG_BUILDER_IMAGE}:${GOLANG_BUILDER_TAG} \
#     /bin/sh scripts/generate_bin.sh
# 
# change permission
# chmod +x ./mycontroller

# get backend branch details
BACKEND_BRANCH=`git rev-parse --abbrev-ref HEAD`

# build web console
# git submodule update --init --recursive
# git submodule update --remote
# cd console-web
# git checkout $BACKEND_BRANCH  # sync with backend branch for webconsole
# yarn install
# CI=false yarn build
# cd ../

# disable UI build for a while
mkdir -p console-web/build

# build conatiner images
docker buildx build --push --progress=plain --platform linux/amd64 --file docker/all-in-one.Dockerfile --tag ${IMAGE_ALL_IN_ONE}:${IMAGE_TAG} .
docker buildx build --push --progress=plain --platform linux/amd64 --file docker/core.Dockerfile --tag ${IMAGE_CORE}:${IMAGE_TAG} .
docker buildx build --push --progress=plain --platform linux/amd64 --file docker/gateway.Dockerfile --tag ${IMAGE_GATEWAY}:${IMAGE_TAG} .

# push images to registry
# docker push ${IMAGE_ALL_IN_ONE}:${IMAGE_TAG}
# docker push ${IMAGE_CORE}:${IMAGE_TAG}
# docker push ${IMAGE_GATEWAY}:${IMAGE_TAG}
