#!/bin/bash

# container registry
REGISTRY='quay.io/mycontroller-org'
IMAGE_NAME="${REGISTRY}/mycontroller"
IMAGE_TAG="2.0-master"  # application tag

# alpine golang builder image
GOLANG_BUILDER_IMAGE="quay.io/mycontroller-org/golang"
GOLANG_BUILDER_TAG="1.15.6-alpine3.12"

# debug lines
echo $PWD
ls -alh
git branch

# build application inside continer
docker run --rm \
    -v "$PWD"/:/usr/src/mycontroller -w /usr/src/mycontroller \
    ${GOLANG_BUILDER_IMAGE}:${GOLANG_BUILDER_TAG} \
    /bin/sh scripts/generate_bin.sh

# change permission
chmod +x ./mycontroller

# get backend branch details
BACKEND_BRANCH=`git rev-parse --abbrev-ref HEAD`

# build web console
git submodule update --init --recursive
git submodule update --remote
cd console-web
git checkout $BACKEND_BRANCH  # sync with backend branch for webconsole
yarn install
CI=false yarn build
cd ../

# build conatiner image
docker build -f docker/Dockerfile -t ${IMAGE_NAME}:${IMAGE_TAG} .

# push image to registry
docker push ${IMAGE_NAME}:${IMAGE_TAG}
