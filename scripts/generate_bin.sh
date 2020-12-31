BUILD_DATE=`date -u +'%Y-%m-%dT%H:%M:%S%:z'`
GIT_BRANCH=`git rev-parse --abbrev-ref HEAD`
GIT_SHA=`git rev-parse HEAD`
GIT_SHA_SHORT=`git rev-parse --short HEAD`
VERSION_PKG="github.com/mycontroller-org/backend/v2/pkg/version"

LD_FLAGS="-X $VERSION_PKG.version=$GIT_BRANCH -X $VERSION_PKG.buildDate=$BUILD_DATE -X $VERSION_PKG.gitCommit=$GIT_SHA"

# build all-in-one image
go build -v -o mycontroller-all-in-one -ldflags "$LD_FLAGS" cmd/all-in-one/main.go

# build core image
go build -v -o mycontroller-core -ldflags "$LD_FLAGS" cmd/core/main.go

# build gateway image
go build -v -o mycontroller-gateway -ldflags "$LD_FLAGS" cmd/gateway/main.go