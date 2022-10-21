.PHONY: default
default: build ;

SHELL := /bin/bash -o pipefail

# the rationale for using both `git describe` and `git rev-parse` is because
# when CI builds the application it can be based on a git tag, so this ensures
# the output is consistent across environments.
VERSION ?= $(shell git describe --tags 2>/dev/null || git rev-parse --short HEAD)

# Allows overriding go executable.
GO_BIN ?= go

# Enables support for tools such as https://github.com/rakyll/gotest
TEST_COMMAND ?= $(GO_BIN) test

# The compute tests can sometimes exceed the default 10m limit.
TEST_ARGS ?= -timeout 15m ./...

CLI_ENV ?= "development"

GOHOSTOS ?= $(shell go env GOHOSTOS || echo unknown)
GOHOSTARCH ?= $(shell go env GOHOSTARCH || echo unknown)

ifeq ($(OS), Windows_NT)
	SHELL = cmd.exe
	.SHELLFLAGS = /c
	GO_FILES = $(shell where /r pkg *.go)
	GO_FILES += $(shell where /r cmd *.go)
	CONFIG_SCRIPT = scripts\config.sh
else
	GO_FILES = $(shell find cmd pkg -type f -name '*.go')
	CONFIG_SCRIPT = scripts/config.sh
endif

# You can pass flags to goreleaser via GORELEASER_ARGS
# --skip-validate will skip the checks
# --rm-dist will save you deleting the dist dir
# --single-target will be quicker and only build for your os & architecture
# e.g.
# make fastly GORELEASER_ARGS="--skip-validate --rm-dist"
fastly: dependencies $(GO_FILES)
	@GOHOSTOS="${GOHOSTOS}" GOHOSTARCH="${GOHOSTARCH}" goreleaser build ${GORELEASER_ARGS}

# Useful for attaching a debugger such as https://github.com/go-delve/delve
debug:
	@$(GO_BIN) build -gcflags="all=-N -l" $(GO_ARGS) -o "fastly" ./cmd/fastly

.PHONY: config
config:
	@$(CONFIG_SCRIPT)

.PHONY: all
all: config dependencies tidy fmt vet staticcheck gosec test build install

# Update CI tools used by ./.github/workflows/pr_test.yml
.PHONY: dependencies
dependencies:
	$(GO_BIN) install github.com/securego/gosec/v2/cmd/gosec@latest
	$(GO_BIN) install honnef.co/go/tools/cmd/staticcheck@latest
	$(GO_BIN) install github.com/mgechev/revive@latest
	$(GO_BIN) install github.com/goreleaser/goreleaser@v1.9.2

# Clean up Go modules file.
.PHONY: tidy
tidy:
	$(GO_BIN) mod tidy

# Run formatter.
.PHONY: fmt
fmt:
	@echo gofmt -l ./{cmd,pkg}
	@eval "bash -c 'F=\$$(gofmt -l ./{cmd,pkg}) ; if [[ \$$F ]] ; then echo \$$F ; exit 1 ; fi'"

# Run static analysis.
.PHONY: vet
vet:
	$(GO_BIN) vet ./{cmd,pkg}/...

# Run linter.
.PHONY: revive
revive:
	revive ./...

# Run security vulnerability checker.
.PHONY: gosec
gosec:
	gosec -quiet -exclude=G104 ./{cmd,pkg}/...

# Run third-party static analysis.
.PHONY: staticcheck
staticcheck:
	staticcheck ./{cmd,pkg}/...

# Run tests
.PHONY: test
test: config
	@$(TEST_COMMAND) -race $(TEST_ARGS)

# Compile program.
#
# GO_ARGS allows for passing additional arguments.
# e.g. make build GO_ARGS='--ldflags "-s -w"'
.PHONY: build
build: config
	$(GO_BIN) build $(GO_ARGS) ./cmd/fastly

# Compile and install program.
#
# GO_ARGS allows for passing additional arguments.
.PHONY: install
install: config
	$(GO_BIN) install $(GO_ARGS) ./cmd/fastly

# Scaffold a new CLI command from template files.
.PHONY: scaffold
scaffold:
	@$(shell pwd)/scripts/scaffold.sh $(CLI_PACKAGE) $(CLI_COMMAND) $(CLI_API)

# Scaffold a new CLI 'category' command from template files.
.PHONY: scaffold-category
scaffold-category:
	@$(shell pwd)/scripts/scaffold-category.sh $(CLI_CATEGORY) $(CLI_CATEGORY_COMMAND) $(CLI_PACKAGE) $(CLI_COMMAND) $(CLI_API)

.PHONY: clean
