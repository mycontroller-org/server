#!/bin/bash

# this script used to generate binary files
# should be executed from the root locations of the repository

BUILD_DATE=`date -u +'%Y-%m-%dT%H:%M:%S%:z'`
GIT_BRANCH=`git rev-parse --abbrev-ref HEAD`
GIT_SHA=`git rev-parse HEAD`
GIT_SHA_SHORT=`git rev-parse --short HEAD`
VERSION_PKG="github.com/mycontroller-org/backend/v2/pkg/version"

LD_FLAGS="-X $VERSION_PKG.version=$GIT_BRANCH -X $VERSION_PKG.buildDate=$BUILD_DATE -X $VERSION_PKG.gitCommit=$GIT_SHA"

# different architectures
LINUX_ARCHITECTURES="arm arm64 386 amd64"
WINDOWS_ARCHITECTURES="386 amd64"

build_binaries() {
  os=$1
  arch=$2
  mkdir -p binaries
  GOOS=${os} GOARCH=${arch} go build -o binaries/mycontroller-all-in-one_${os}_${arch} -ldflags "$LD_FLAGS" cmd/all-in-one/main.go
  GOOS=${os} GOARCH=${arch} go build -o binaries/mycontroller-core_${os}_${arch} -ldflags "$LD_FLAGS" cmd/core/main.go
  GOOS=${os} GOARCH=${arch} go build -o binaries/mycontroller-gateway_${os}_${arch} -ldflags "$LD_FLAGS" cmd/gateway/main.go
}

# download dependencies
go mod tidy
 
# compile for linux
for arch in ${LINUX_ARCHITECTURES}
do
  build_binaries "linux" ${arch}
done


# compile for windows
for arch in ${WINDOWS_ARCHITECTURES}
do
  build_binaries "windows" ${arch}
done
