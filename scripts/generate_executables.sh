#!/bin/bash

# this script used to generate binary files
# should be executed from the root locations of the repository


source ./scripts/version.sh

BUILD_DIR=builds
BINARY_DIR=binary
rm -rf "${BUILD_DIR}"
mkdir -p ${BUILD_DIR}/${BINARY_DIR}

# generate UI builds
if [ "${BUILD_UI}" = true ] ; then
  ./scripts/build_web_console.sh
fi

# pack the UI into the server binary (extract-on-start)
go run ./scripts/cmd/pack_web_console
if [ $? -ne 0 ]; then
  echo "failed to pack web console assets"
  exit 1
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

  if [ ${COMPONENT_NAME} = "server" ]; then
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

compile() {
  local goos=$1
  local goarch=$2
  local goarm=$3
  local tags=$4
  local outfile=$5
  local source=$6

  echo "building ${outfile} (GOOS=${goos} GOARCH=${goarch} GOARM=${goarm} tags=${tags})"
  local -a build_cmd=(env CGO_ENABLED=0 GOOS=${goos} GOARCH=${goarch} GOARM=${goarm} go build -trimpath)
  if [ -n "${tags}" ]; then
    build_cmd+=(-tags="${tags}")
  fi
  build_cmd+=(-o "${outfile}" -ldflags "$LD_FLAGS" "${source}")
  if ! "${build_cmd[@]}"; then
    echo "an error has occurred. aborting the build process"
    exit 1
  fi
}

# host-built binaries copied into alpine images
copy_docker_binary() {
  local goos=$1
  local goarch=$2
  local goarm=$3
  local component=$4
  local src=$5

  if [ "${goos}" != "linux" ]; then
    return
  fi

  local dir
  if [ "${goarch}" = "arm" ]; then
    dir="${BUILD_DIR}/${BINARY_DIR}/linux-armv${goarm:-7}"
  else
    dir="${BUILD_DIR}/${BINARY_DIR}/linux-${goarch}"
  fi
  mkdir -p "${dir}"
  cp "${src}" "${dir}/mycontroller-${component}"
  chmod +x "${dir}/mycontroller-${component}"
}

# platforms to build archives
PLATFORMS=("linux/arm" "linux/arm64" "linux/386" "linux/amd64" "windows/386" "windows/amd64")

# compile
for platform in "${PLATFORMS[@]}"
do
  platform_raw=(${platform//\// })
  GOOS=${platform_raw[0]}
  GOARCH=${platform_raw[1]}
  GOARM=""
  if [ "${GOARCH}" = "arm" ]; then
    GOARM="7"
  fi
  package_server="mycontroller-server-${GOOS}-${GOARCH}"
  package_gateway="mycontroller-gateway-${GOOS}-${GOARCH}"
  package_handler="mycontroller-handler-${GOOS}-${GOARCH}"
  package_client="myc-${GOOS}-${GOARCH}"

  compile ${GOOS} ${GOARCH} "${GOARM}" "server,embedui" ${BUILD_DIR}/${BINARY_DIR}/${package_server} cmd/component/server/main.go
  compile ${GOOS} ${GOARCH} "${GOARM}" "standalone" ${BUILD_DIR}/${BINARY_DIR}/${package_gateway} cmd/component/gateway/main.go
  compile ${GOOS} ${GOARCH} "${GOARM}" "standalone" ${BUILD_DIR}/${BINARY_DIR}/${package_handler} cmd/component/handler/main.go
  compile ${GOOS} ${GOARCH} "${GOARM}" "" ${BUILD_DIR}/${BINARY_DIR}/${package_client} cmd/client/main.go

  copy_docker_binary ${GOOS} ${GOARCH} "${GOARM}" server ${BUILD_DIR}/${BINARY_DIR}/${package_server}
  copy_docker_binary ${GOOS} ${GOARCH} "${GOARM}" gateway ${BUILD_DIR}/${BINARY_DIR}/${package_gateway}
  copy_docker_binary ${GOOS} ${GOARCH} "${GOARM}" handler ${BUILD_DIR}/${BINARY_DIR}/${package_handler}

  FILE_EXTENSION=""
  if [ $GOOS = "windows" ]; then
    FILE_EXTENSION='.exe'
  fi

  package mycontroller-server-${VERSION}-${GOOS}-${GOARCH} "server" ${package_server} ${FILE_EXTENSION}
  package mycontroller-gateway-${VERSION}-${GOOS}-${GOARCH} "gateway" ${package_gateway} ${FILE_EXTENSION}
  package mycontroller-handler-${VERSION}-${GOOS}-${GOARCH} "handler" ${package_handler} ${FILE_EXTENSION}
  package mycontroller-client-${VERSION}-${GOOS}-${GOARCH} "client" ${package_client} ${FILE_EXTENSION}
done

# extra linux/armv6 binaries for container images (not archived)
for component_spec in "server:server,embedui:cmd/component/server/main.go" "gateway:standalone:cmd/component/gateway/main.go" "handler:standalone:cmd/component/handler/main.go"; do
  IFS=':' read -r component tags source <<< "${component_spec}"
  outfile=${BUILD_DIR}/${BINARY_DIR}/mycontroller-${component}-linux-armv6
  compile linux arm 6 "${tags}" "${outfile}" "${source}"
  copy_docker_binary linux arm 6 "${component}" "${outfile}"
done

