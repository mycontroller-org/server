#!/bin/bash

# this script used to generate binary files
# should be executed from the root locations of the repository


source ./scripts/version.sh

BUILD_DIR=builds
BINARY_DIR=binary
# clean builds directory
rm ${BUILD_DIR}/* -rf


# create directories
mkdir -p ${BUILD_DIR}/${BINARY_DIR}

# generate UI builds
if [ "${BUILD_UI}" = true ] ; then
  ./scripts/build_web_console.sh
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
  if [[ "${COMPONENT_NAME}" == "client" ]]; then
      cp ${BUILD_DIR}/${BINARY_DIR}/${BINARY_FILE} ${PACKAGE_STAGING_DIR}/myc${FILE_EXTENSION}
  else
    cp ${BUILD_DIR}/${BINARY_DIR}/${BINARY_FILE} ${PACKAGE_STAGING_DIR}/mycontroller-${COMPONENT_NAME}${FILE_EXTENSION}
  fi

  # config file name
  local CONFIG_FILE=${COMPONENT_NAME}.yaml

  # include web console
  if [ ${COMPONENT_NAME} = "server" ]; then
    cp web-console/build ${PACKAGE_STAGING_DIR}/web_console -r
    CONFIG_FILE="mycontroller.yaml"    
  fi

  if [[ "${COMPONENT_NAME}" != "client" ]]; then
    # copy sample config file
    cp resources/sample-binary-${COMPONENT_NAME}.yaml ${PACKAGE_STAGING_DIR}/${CONFIG_FILE}
    # copy start/stop script
    cp resources/control-scripts/mcctl-${COMPONENT_NAME}.sh ${PACKAGE_STAGING_DIR}/mcctl.sh
    # copy readme text
    cp resources/control-scripts/README.txt ${PACKAGE_STAGING_DIR}/README.txt
  fi
  
  # copy license
  cp LICENSE ${PACKAGE_STAGING_DIR}/LICENSE.txt

  if [[ ${PACKAGE_STAGING_DIR} =~ "windows" ]]; then
    ARCHIVE_NAME="${PACKAGE_STAGING_DIR}.zip"
    zip -r ${BUILD_DIR}/${ARCHIVE_NAME} ${PACKAGE_STAGING_DIR}
  else
    ARCHIVE_NAME="${PACKAGE_STAGING_DIR}.tar.gz"
    tar -czf ${BUILD_DIR}/${ARCHIVE_NAME} ${PACKAGE_STAGING_DIR}
  fi

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
  package_server="mycontroller-server-${GOOS}-${GOARCH}"
  package_gateway="mycontroller-gateway-${GOOS}-${GOARCH}"
  package_handler="mycontroller-handler-${GOOS}-${GOARCH}"
  package_client="myc"

  # to use embed web assets use tag "web"
  # embed assets takes extra ~40 MiB when running
  env GOOS=${GOOS} GOARCH=${GOARCH} go build -tags=server -o ${BUILD_DIR}/${BINARY_DIR}/${package_server} -ldflags "$LD_FLAGS" cmd/component/server/main.go
  build_status=$?
  if [ $build_status -ne 0 ]; then
      echo "an error has occurred. aborting the build process, status:${build_status}"
      exit $status
  fi
  env GOOS=${GOOS} GOARCH=${GOARCH} go build -tags=standalone -o ${BUILD_DIR}/${BINARY_DIR}/${package_gateway} -ldflags "$LD_FLAGS" cmd/component/gateway/main.go
  build_status=$?
  if [ $build_status -ne 0 ]; then
      echo "an error has occurred. aborting the build process, status:${build_status}"
      exit $status
  fi
  env GOOS=${GOOS} GOARCH=${GOARCH} go build -tags=standalone -o ${BUILD_DIR}/${BINARY_DIR}/${package_handler} -ldflags "$LD_FLAGS" cmd/component/handler/main.go
  build_status=$?
  if [ $build_status -ne 0 ]; then
      echo "an error has occurred. aborting the build process, status:${build_status}"
      exit $status
  fi
  env GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${BUILD_DIR}/${BINARY_DIR}/${package_client} -ldflags "$LD_FLAGS" cmd/client/main.go
  build_status=$?
  if [ $build_status -ne 0 ]; then
      echo "an error has occurred. aborting the build process, status:${build_status}"
      exit $status
  fi

  FILE_EXTENSION=""
  if [ $GOOS = "windows" ]; then
    FILE_EXTENSION='.exe'
  fi

  package mycontroller-server-${VERSION}-${GOOS}-${GOARCH} "server" ${package_server} ${FILE_EXTENSION}
  package mycontroller-gateway-${VERSION}-${GOOS}-${GOARCH} "gateway" ${package_gateway} ${FILE_EXTENSION}
  package mycontroller-handler-${VERSION}-${GOOS}-${GOARCH} "handler" ${package_handler} ${FILE_EXTENSION}
  package mycontroller-client-${VERSION}-${GOOS}-${GOARCH} "client" ${package_client} ${FILE_EXTENSION}
done

