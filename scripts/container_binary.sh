#!/bin/sh

source ./scripts/version.sh

TARGET_BINARY=${TARGET_BUILD:-server}

if [[ "${TARGET_BINARY}" == "gateway" ]]; then # build gateway binary
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -trimpath -tags=standalone -o mycontroller-gateway -ldflags "$LD_FLAGS" cmd/component/gateway/main.go

elif [[ "${TARGET_BINARY}" == "handler" ]]; then # build handler binary
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -trimpath -tags=standalone -o mycontroller-handler -ldflags "$LD_FLAGS" cmd/component/handler/main.go

elif [[ "${TARGET_BINARY}" == "client" ]]; then # build client binary
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -trimpath -o myc -ldflags "$LD_FLAGS" cmd/client/main.go

else
  if [ ! -f web-console/build/index.html ]; then
    echo "web-console/build/index.html not found; build the console before the server image" >&2
    exit 1
  fi
  go run ./scripts/cmd/pack_web_console
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -trimpath -tags=server,embedui -o mycontroller-server -ldflags "$LD_FLAGS" cmd/component/server/main.go
fi
