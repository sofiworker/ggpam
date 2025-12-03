BIN_DIR := bin
CLI_BINARY := $(BIN_DIR)/ggpam
PAM_SO := $(BIN_DIR)/pam_ggpam.so
PAM_HEADER := $(BIN_DIR)/pam_ggpam.h
GOFMT_FILES := $(shell find . -name '*.go' -not -path './dist/*' -not -path './bin/*' -not -path './vendor/*')

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION ?= $(shell go env GOVERSION)
LD_FLAGS := -X ggpam/pkg/version.Version=$(VERSION) \
	-X ggpam/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X ggpam/pkg/version.BuildDate=$(BUILD_DATE) \
	-X ggpam/pkg/version.GoVersion=$(GO_VERSION)
LD_FLAGS += $(EXTRA_LD_FLAGS)

.PHONY: build test fmt lint clean deps package deb rpm rpm-debug pam cli

build: $(CLI_BINARY) $(PAM_SO)

cli: $(CLI_BINARY)

pam: $(PAM_SO)

$(CLI_BINARY):
	@mkdir -p $(BIN_DIR)
	@echo "==> go build (CLI)"
	go build -ldflags "$(LD_FLAGS)" -o $(CLI_BINARY) ./cmd/cli

$(PAM_SO):
	@mkdir -p $(BIN_DIR)
	@echo "==> go build (PAM shared library)"
	go build -buildmode=c-shared -ldflags "$(LD_FLAGS)" -o $(PAM_SO) ./cmd/pam

test:
	go test ./...

fmt:
	gofmt -w $(GOFMT_FILES)

lint:
	go vet ./...

deps:
	./scripts/check_deps.sh

clean:
	rm -rf bin dist

deb:
	./scripts/build_deb.sh

rpm:
	./scripts/build_rpm.sh

package: deb rpm
