#!/bin/bash

# docker registry
DOCKER_ORG='quay.io/mycontroller-org'
DOCKER_REPO="${DOCKER_ORG}/mycontroller"

# alpine golang builder image tag
GOLANG_BUILDER_TAG="1.15.0-alpine3.12"

# tag version
TAG="2.0-master"

# build go project
# go build ../main.go
docker run --rm -v \
    "$PWD"/..:/usr/src/mycontroller -w /usr/src/mycontroller \
    golang:${GOLANG_BUILDER_TAG} \
    go build -v -o docker/mycontroller cmd/main.go

# change permission
chmod +x ./mycontroller

# copy default templates
cp ../resources/default.yaml ./default.yaml

# build web console
cd ../
git submodule update --init --recursive
git submodule update --remote
cd console-web
yarn install
yarn build
cd ../docker

# copy UI files
cp ../console-web/build ./ui -r

# build image
docker build -t ${DOCKER_REPO}:${TAG} .

# push image to registry
docker push ${DOCKER_REPO}:${TAG}

# remove application bin file and UI files
rm ./mycontroller -rf
rm ./default.config -rf
rm ./ui -rf
