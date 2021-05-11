#!/bin/bash

# this script used to generate binary files
# should be executed from the root locations of the repository

BUILD_DIR=builds
# clean builds directory
rm ${BUILD_DIR}/* -rf

BUILD_DATE=`date -u +'%Y-%m-%dT%H:%M:%S%:z'`
GIT_BRANCH=`git rev-parse --abbrev-ref HEAD`
GIT_SHA=`git rev-parse HEAD`
GIT_SHA_SHORT=`git rev-parse --short HEAD`
VERSION_PKG="github.com/mycontroller-org/backend/v2/pkg/version"

LD_FLAGS="-X $VERSION_PKG.version=$GIT_BRANCH -X $VERSION_PKG.buildDate=$BUILD_DATE -X $VERSION_PKG.gitCommit=$GIT_SHA"

build_binaries() {
  os=$1
  arch=$2
  mkdir -p ${BUILD_DIR}
  GOOS=${os} GOARCH=${arch} go build -o ${BUILD_DIR}/mycontroller-all-in-one_${GIT_BRANCH}_${os}-${arch} -ldflags "$LD_FLAGS" cmd/all-in-one/main.go
  GOOS=${os} GOARCH=${arch} go build -o ${BUILD_DIR}/mycontroller-core_${GIT_BRANCH}_${os}-${arch} -ldflags "$LD_FLAGS" cmd/core/main.go
  GOOS=${os} GOARCH=${arch} go build -o ${BUILD_DIR}/mycontroller-gateway_${GIT_BRANCH}_${os}_${arch} -ldflags "$LD_FLAGS" cmd/gateway/main.go
}

# download dependencies
go mod tidy

# different architectures
LINUX_ARCHITECTURES="arm arm64 386 amd64"
WINDOWS_ARCHITECTURES="386 amd64"
 
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

# generate UI builds
if [ "$BUILD_UI" = true ] ; then
  ./scripts/build_ui.sh
fi