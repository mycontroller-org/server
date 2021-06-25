#!/bin/sh

source ./scripts/version.sh

TARGET_BINARY=${TARGET_BUILD:-server}

if [[ "${TARGET_BINARY}" == "gateway" ]]; then # build gateway binary
  GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -tags=standalone -o mycontroller-gateway -ldflags "$LD_FLAGS" cmd/gateway/main.go

elif [[ "${TARGET_BINARY}" == "handler" ]]; then # build handler binary
  GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -tags=standalone -o mycontroller-handler -ldflags "$LD_FLAGS" cmd/handler/main.go

else # build server binary
  GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -tags=server,web -o mycontroller-server -ldflags "$LD_FLAGS" cmd/server/main.go

fi
