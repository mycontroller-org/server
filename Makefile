# Local MyController build
#
#   make              # server + gateway + handler + client
#   make server
#   make test

BIN_DIR     ?= builds
VERSION_PKG := github.com/mycontroller-org/server/v2/pkg/version
VERSION     ?= $(shell grep '^server=' versions.txt | cut -d= -f2)-devel
GIT_COMMIT  ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_DATE  ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS     := -X $(VERSION_PKG).version=$(VERSION) -X $(VERSION_PKG).buildDate=$(BUILD_DATE) -X $(VERSION_PKG).gitCommit=$(GIT_COMMIT)

.PHONY: help all build server gateway handler client test clean

help:
	@echo "Targets:"
	@echo "  make / make build   build all binaries into $(BIN_DIR)/"
	@echo "  make server         $(BIN_DIR)/mycontroller-server"
	@echo "  make gateway        $(BIN_DIR)/mycontroller-gateway"
	@echo "  make handler        $(BIN_DIR)/mycontroller-handler"
	@echo "  make client         $(BIN_DIR)/myc"
	@echo "  make test           go test ./..."
	@echo "  make clean          remove $(BIN_DIR)/"

all build: server gateway handler client

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

server: $(BIN_DIR)
	go build -trimpath -tags=server -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/mycontroller-server ./cmd/component/server

gateway: $(BIN_DIR)
	go build -trimpath -tags=standalone -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/mycontroller-gateway ./cmd/component/gateway

handler: $(BIN_DIR)
	go build -trimpath -tags=standalone -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/mycontroller-handler ./cmd/component/handler

client: $(BIN_DIR)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/myc ./cmd/client

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)
