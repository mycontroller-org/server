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

# versions.txt is the next release to cut (x.y.z). Do not reuse VERSION:
# that is the build string from version.sh (often x.y.z-devel).
RELEASE_VERSION ?= $(shell awk -F= '/^server=/{print $$2}' versions.txt)
ifeq ($(origin VERSION),command line)
NEXT_FROM := $(VERSION)
else
NEXT_FROM := $(RELEASE_VERSION)
endif
NEXT_VERSION := $(shell printf '%s\n' '$(NEXT_FROM)' | awk -F. '{if (NF==3 && $$1~/^[0-9]+$$/ && $$2~/^[0-9]+$$/ && $$3~/^[0-9]+$$/) printf "%d.%d.%d", $$1, $$2, $$3+1}')

.PHONY: help all build server gateway handler client web-console test clean setup-release next-version

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
	@echo "  make next-version   print the next patch after versions.txt (or VERSION=x.y.z)"
	@echo "  make setup-release [VERSION=x.y.z]   sync main, bump versions.txt, open release PR"

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

next-version:
	@if [ -z "$(NEXT_VERSION)" ]; then echo "could not compute next version from '$(NEXT_FROM)'" >&2; exit 1; fi
	@echo $(NEXT_VERSION)

setup-release:
ifeq ($(origin VERSION),command line)
	./scripts/setup_release.sh "$(VERSION)"
else
	./scripts/setup_release.sh "$(RELEASE_VERSION)"
endif
