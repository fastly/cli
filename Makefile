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
TEST_ARGS ?= -timeout 15m ./{cmd,pkg}/...

CLI_ENV ?= "development"

GOHOSTOS ?= $(shell go env GOHOSTOS || echo unknown)
GOHOSTARCH ?= $(shell go env GOHOSTARCH || echo unknown)
GO_FILES = $(shell find cmd pkg -type f -name '*.go')

# You can pass flags to goreleaser via GORELEASER_ARGS
# --skip-validate will skip the checks
# --rm-dist will save you deleting the dist dir
# --single-target will be quicker and only build for your os & architecture
# e.g.
# make fastly GORELEASER_ARGS="--skip-validate --rm-dist"
fastly: dependencies $(GO_FILES)
	@GOHOSTOS="${GOHOSTOS}" GOHOSTARCH="${GOHOSTARCH}" goreleaser build ${GORELEASER_ARGS}


# useful for attaching a debugger such as https://github.com/go-delve/delve
debug:
	@$(GO_BIN) build -gcflags="all=-N -l" $(GO_ARGS) -o "fastly" ./cmd/fastly

.PHONY: all
all: dependencies config tidy fmt vet staticcheck gosec test build install

# update goreleaser inline with the release GHA workflow
.PHONY: dependencies
dependencies:
	$(GO_BIN) install github.com/securego/gosec/v2/cmd/gosec@latest
	$(GO_BIN) install honnef.co/go/tools/cmd/staticcheck@latest
	$(GO_BIN) install github.com/mgechev/revive@latest
	$(GO_BIN) install github.com/goreleaser/goreleaser@v1.9.2

.PHONY: tidy
tidy:
	$(GO_BIN) mod tidy

.PHONY: fmt
fmt:
	@echo gofmt -l ./{cmd,pkg}
	@eval "bash -c 'F=\$$(gofmt -l ./{cmd,pkg}) ; if [[ \$$F ]] ; then echo \$$F ; exit 1 ; fi'"

.PHONY: vet
vet:
	$(GO_BIN) vet ./{cmd,pkg}/...

.PHONY: revive
revive:
	revive ./...

.PHONY: gosec
gosec:
	gosec -quiet -exclude=G104 ./{cmd,pkg}/...

.PHONY: staticcheck
staticcheck:
	staticcheck ./{cmd,pkg}/...

.PHONY: test
test: config
	@$(TEST_COMMAND) -race $(TEST_ARGS)

# GO_ARGS allows for passing additional args to the build e.g.
# make build GO_ARGS='--ldflags "-s -w"'
.PHONY: build
build: config
	$(GO_BIN) build $(GO_ARGS) ./cmd/fastly

# GO_ARGS allows for passing additional args to the build e.g.
.PHONY: install
install: config
	$(GO_BIN) install $(GO_ARGS) ./cmd/fastly

.PHONY: changelog
changelog:
	@$(shell pwd)/scripts/changelog.sh

.PHONY: release-changelog
release-changelog:
	@$(shell pwd)/scripts/release-changelog.sh

.PHONY: config
config:
	@curl -so cmd/fastly/static/config.toml https://developer.fastly.com/api/internal/cli-config

.PHONY: scaffold
scaffold:
	@$(shell pwd)/scripts/scaffold.sh $(CLI_PACKAGE) $(CLI_COMMAND) $(CLI_API)

.PHONY: scaffold-category
scaffold-category:
	@$(shell pwd)/scripts/scaffold-category.sh $(CLI_CATEGORY) $(CLI_CATEGORY_COMMAND) $(CLI_PACKAGE) $(CLI_COMMAND) $(CLI_API)
