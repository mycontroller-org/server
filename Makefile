# Local MyController build
#
#   make              # server + gateway + handler + client
#   make server
#   make test

BIN_DIR     ?= builds
VERSION_PKG := github.com/mycontroller-org/server/v2/pkg/version
# Same rules as scripts/version.sh: trailing x.y.z from the current tag/branch,
# otherwise versions.txt with a -devel suffix.
VERSION     ?= $(shell bash -c '. ./scripts/version.sh && echo $$VERSION')
GIT_COMMIT  ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_DATE  ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS     := -s -w -X $(VERSION_PKG).version=$(VERSION) -X $(VERSION_PKG).buildDate=$(BUILD_DATE) -X $(VERSION_PKG).gitCommit=$(GIT_COMMIT)

.PHONY: help all build server gateway handler client web-console test clean setup-release

help:
	@echo "Targets:"
	@echo "  make / make build   build all binaries into $(BIN_DIR)/"
	@echo "  make server         $(BIN_DIR)/mycontroller-server"
	@echo "  make gateway        $(BIN_DIR)/mycontroller-gateway"
	@echo "  make handler        $(BIN_DIR)/mycontroller-handler"
	@echo "  make client         $(BIN_DIR)/myc"
	@echo "  make web-console    production UI + pack zip for embedui builds"
	@echo "  make test           go test ./..."
	@echo "  make clean          remove $(BIN_DIR)/"
	@echo "  make setup-release VERSION=x.y.z NEXT_VERSION=x.y.z   sync main, set next versions.txt, open release PR"

all build: server gateway handler client

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

SERVER_TAGS := server
ifneq ($(wildcard pkg/http_router/web-console/assets/web_console.zip),)
SERVER_TAGS := server,embedui
endif

server: $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -tags=$(SERVER_TAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/mycontroller-server ./cmd/component/server

gateway: $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -tags=standalone -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/mycontroller-gateway ./cmd/component/gateway

handler: $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -tags=standalone -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/mycontroller-handler ./cmd/component/handler

client: $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/myc ./cmd/client

web-console:
	./scripts/build_web_console.sh
	go run ./scripts/cmd/pack_web_console

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)

setup-release:
	@if [ "$(origin VERSION)" != "command line" ] || [ "$(origin NEXT_VERSION)" != "command line" ]; then \
		echo "usage: make setup-release VERSION=x.y.z NEXT_VERSION=x.y.z" >&2; exit 1; fi
	./scripts/setup_release.sh "$(VERSION)" "$(NEXT_VERSION)"
