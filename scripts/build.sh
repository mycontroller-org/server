#!/bin/bash

# docker registry
DOCKER_ORG='quay.io/mycontroller-org'
DOCKER_REPO="${DOCKER_ORG}/mycontroller"

# alpine golang builder image tag
GOLANG_BUILDER_TAG="1.15.0-alpine3.12"

# tag version
TAG="2.0-master"

# debug lines
echo $PWD
ls -alh
git branch

# build go project
# go build ../main.go
docker run --rm -v \
    "$PWD"/:/usr/src/mycontroller -w /usr/src/mycontroller \
    golang:${GOLANG_BUILDER_TAG} \
    go build -v -o mycontroller cmd/main.go

# change permission
chmod +x ./mycontroller

# build web console
git submodule update --init --recursive
git submodule update --remote
cd console-web
yarn install
CI=false yarn build
cd ../

# build image
docker build -f docker/Dockerfile -t ${DOCKER_REPO}:${TAG} .

# push image to registry
docker push ${DOCKER_REPO}:${TAG}
