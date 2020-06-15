SHELL := /bin/bash -o pipefail

VERSION ?= $(shell git describe --tags 2>/dev/null || git rev-parse --short HEAD)
LDFLAGS = -ldflags "\
 -X 'github.com/fastly/cli/pkg/version.AppVersion=${VERSION}' \
 -X 'github.com/fastly/cli/pkg/version.GitRevision=$(shell git rev-parse --short HEAD || echo unknown)' \
 -X 'github.com/fastly/cli/pkg/version.GoVersion=$(shell go version)' \
 "

fastly:
	@go build -trimpath $(LDFLAGS) -o "$@" ./cmd/fastly

.PHONY: all
all: tidy fmt vet staticcheck lint gosec test build install

.PHONY: dependencies
dependencies:
	go get -v -u github.com/securego/gosec/cmd/gosec
	go get -v -u honnef.co/go/tools/cmd/staticcheck
	go get -v -u golang.org/x/lint/golint

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

.PHONY: lint
lint:
	golint ./{cmd,pkg}/...

.PHONY: fossa
fossa:
	fossa analyze -e $$FOSSA_API_ENDPOINT
	fossa test -e $$FOSSA_API_ENDPOINT

.PHONY: test
test:
	go test -race ./{cmd,pkg}/...

.PHONY: build
build:
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
