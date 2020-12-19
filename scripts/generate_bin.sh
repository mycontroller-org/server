BUILD_DATE=`date -u +'%Y-%m-%dT%H:%M:%S%:z'`
GIT_BRANCH=`git rev-parse --abbrev-ref HEAD`
GIT_SHA=`git rev-parse HEAD`
GIT_SHA_SHORT=`git rev-parse --short HEAD`
VERSION_PKG="github.com/mycontroller-org/backend/v2/pkg/version"

LD_FLAGS="-X $VERSION_PKG.version=$GIT_BRANCH -X $VERSION_PKG.buildDate=$BUILD_DATE -X $VERSION_PKG.gitCommit=$GIT_SHA"


go build -v -o mycontroller -ldflags "$LD_FLAGS" cmd/main.go