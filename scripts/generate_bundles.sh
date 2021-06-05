#!/bin/bash

# this script used to generate binary files
# should be executed from the root locations of the repository

BUILD_DIR=builds
BINARY_DIR=binary
# clean builds directory
rm ${BUILD_DIR}/* -rf


# create directories
mkdir -p ${BUILD_DIR}/${BINARY_DIR}

BUILD_DATE=`date -u +'%Y-%m-%dT%H:%M:%S%:z'`
GIT_BRANCH=`git rev-parse --abbrev-ref HEAD`
GIT_SHA=`git rev-parse HEAD`
GIT_SHA_SHORT=`git rev-parse --short HEAD`
VERSION_PKG="github.com/mycontroller-org/backend/v2/pkg/version"

LD_FLAGS="-X $VERSION_PKG.version=$GIT_BRANCH -X $VERSION_PKG.buildDate=$BUILD_DATE -X $VERSION_PKG.gitCommit=$GIT_SHA"


# generate UI builds
if [ "${BUILD_UI}" = true ] ; then
  ./scripts/build_ui.sh
fi

# download dependencies
go mod tidy


function package {
  local PACKAGE_STAGING_DIR=$1
  local COMPONENT_NAME=$2
  local BINARY_FILE=$3
  local FILE_EXTENSION=$4

  mkdir -p ${PACKAGE_STAGING_DIR}

  # echo "Package dir: ${PACKAGE_STAGING_DIR}"
  cp ${BUILD_DIR}/${BINARY_DIR}/${BINARY_FILE} ${PACKAGE_STAGING_DIR}/mycontroller-${COMPONENT_NAME}${FILE_EXTENSION}

  # config file name
  local CONFIG_FILE=${COMPONENT_NAME}.yaml

  # include web console
  if [ ${COMPONENT_NAME} = "all-in-one" ] || [ ${COMPONENT_NAME} = "core" ]; then
    CONFIG_FILE="mycontroller.yaml"
    cp console-web/build ${PACKAGE_STAGING_DIR}/web_console -r
  fi

  # copy sample config file
  cp resources/sample-binary-${COMPONENT_NAME}.yaml ${PACKAGE_STAGING_DIR}/${CONFIG_FILE}
  # copy start/stop script
  cp resources/control-scripts/start-${COMPONENT_NAME}.sh ${PACKAGE_STAGING_DIR}/start.sh

  ARCHIVE_NAME="${PACKAGE_STAGING_DIR}.tar.gz"
  # echo "Packaging into: ${ARCHIVE_NAME}"
  tar -czf ${BUILD_DIR}/${ARCHIVE_NAME} ${PACKAGE_STAGING_DIR}
  rm ${PACKAGE_STAGING_DIR} -rf
}

# platforms to build
PLATFORMS=("linux/arm" "linux/arm64" "linux/386" "linux/amd64" "windows/386" "windows/amd64")

# compile
for platform in "${PLATFORMS[@]}"
do
  platform_raw=(${platform//\// })
  GOOS=${platform_raw[0]}
  GOARCH=${platform_raw[1]}
  package_all_in_one="mycontroller-all-in-one-${GOOS}-${GOARCH}"
  package_core="mycontroller-core-${GOOS}-${GOARCH}"
  package_gateway="mycontroller-gateway-${GOOS}-${GOARCH}"

  env GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${BUILD_DIR}/${BINARY_DIR}/${package_all_in_one} -ldflags "$LD_FLAGS" cmd/all-in-one/main.go
  env GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${BUILD_DIR}/${BINARY_DIR}/${package_core} -ldflags "$LD_FLAGS" cmd/core/main.go
  env GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${BUILD_DIR}/${BINARY_DIR}/${package_gateway} -ldflags "$LD_FLAGS" cmd/gateway/main.go
  if [ $? -ne 0 ]; then
    echo 'an error has occurred. aborting the build process'
    exit 1
  fi

  FILE_EXTENSION=""
  if [ $GOOS = "windows" ]; then
    FILE_EXTENSION='.exe'
  fi

  package mycontroller-all-in-one-${GIT_BRANCH}-${GOOS}-${GOARCH} "all-in-one" ${package_all_in_one} ${FILE_EXTENSION}
  package mycontroller-core-${GIT_BRANCH}-${GOOS}-${GOARCH} "core" ${package_core} ${FILE_EXTENSION}
  package mycontroller-gateway-${GIT_BRANCH}-${GOOS}-${GOARCH} "gateway" ${package_gateway} ${FILE_EXTENSION}
done

