SHELL := /bin/bash -o pipefail

# the rationale for using both `git describe` and `git rev-parse` is because
# when CI builds the application it can be based on a git tag, so this ensures
# the output is consistent across environments.
VERSION ?= $(shell git describe --tags 2>/dev/null || git rev-parse --short HEAD)

TESTARGS ?= ./{cmd,pkg}/...

LDFLAGS = -ldflags "\
 -X 'github.com/fastly/cli/pkg/revision.AppVersion=${VERSION}' \
 -X 'github.com/fastly/cli/pkg/revision.GitCommit=$(shell git rev-parse --short HEAD || echo unknown)' \
 -X 'github.com/fastly/cli/pkg/revision.GoVersion=$(shell go version)' \
 "

fastly:
	@go build -trimpath $(LDFLAGS) -o "$@" ./cmd/fastly

# useful for attaching a debugger such as https://github.com/go-delve/delve
debug:
	@go build -gcflags="all=-N -l" -o "fastly" ./cmd/fastly

.PHONY: all
all: config tidy fmt vet staticcheck gosec test build install

.PHONY: dependencies
dependencies:
	go get -v -u github.com/securego/gosec/cmd/gosec
	go get -v -u honnef.co/go/tools/cmd/staticcheck

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	@echo gofmt -l ./{cmd,pkg}
	@eval "bash -c 'F=\$$(gofmt -l ./{cmd,pkg}) ; if [[ \$$F ]] ; then echo \$$F ; exit 1 ; fi'"

.PHONY: vet
vet:
	go vet ./{cmd,pkg}/...

.PHONY: gosec
gosec:
	gosec -quiet -exclude=G104 ./{cmd,pkg}/...

.PHONY: staticcheck
staticcheck:
	staticcheck ./{cmd,pkg}/...

.PHONY: test
test: config
	go test -race $(TESTARGS)

.PHONY: build
build: config
	go build ./cmd/fastly

.PHONY: install
install:
	go install ./cmd/fastly

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
