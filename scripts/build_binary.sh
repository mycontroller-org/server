#!/bin/sh

source ./scripts/version.sh

TARGET_BINARY=${TARGET_BUILD:-all-in-one}

if [[ "${TARGET_BINARY}" == "core" ]]; then  # build core binary
  GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o mycontroller-core -ldflags "$LD_FLAGS" cmd/core/main.go

elif [[ "${TARGET_BINARY}" == "gateway" ]]; then # build gateway binary
  GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o mycontroller-gateway -ldflags "$LD_FLAGS" cmd/gateway/main.go

else # build all-in-one binary
  GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o mycontroller-all-in-one -ldflags "$LD_FLAGS" cmd/all-in-one/main.go

fi
