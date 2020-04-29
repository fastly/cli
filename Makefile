SHELL := /bin/bash -o pipefail

GITHUB_CHANGELOG_GENERATOR := $(shell command -v github_changelog_generator 2> /dev/null)
PREVIOUS_SEMVER_TAG := $(shell git tag | sort -r --version-sort | egrep '^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$$' | head -n2 | tail -n1)
VERSION ?= $(shell git describe --tags 2>/dev/null || git rev-parse --short HEAD)
LDFLAGS = -ldflags "\
 -X 'github.com/fastly/cli/pkg/version.AppVersion=${VERSION}' \
 -X 'github.com/fastly/cli/pkg/version.GitRevision=$(shell git rev-parse --short HEAD || echo unknown)' \
 -X 'github.com/fastly/cli/pkg/version.GoVersion=$(shell go version)' \
 "

fastly:
	@go build -trimpath $(LDFLAGS) -o "$@" ./cmd/fastly

.PHONY: all
all: dependencies fmt vet staticcheck lint gosec test build install

.PHONY: dependencies
dependencies:
	go get -v -u github.com/securego/gosec/cmd/gosec
	go get -v -u honnef.co/go/tools/cmd/staticcheck
	go get -v -u golang.org/x/lint/golint

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

.PHONY:
changelog:
ifndef GITHUB_CHANGELOG_GENERATOR
	$(error "No github_changelog_generator in $$PATH, install via `gem install github_changelog_generator`.")
endif
ifndef CHANGELOG_GITHUB_TOKEN
	@echo ""
	@echo "WARNING: No \$$CHANGELOG_GITHUB_TOKEN in environment, set one to avoid hitting rate limit."
	@echo ""
endif
	github_changelog_generator -u fastly -p cli \
		--no-pr-wo-labels \
		--no-author \
		--enhancement-label "**Enhancements:**" \
		--bugs-label "**Bug fixes:**" \
		--release-url "https://github.com/fastly/cli/releases/tag/%s" \
		--exclude-labels documentation \
		--exclude-tags-regex "v.*-.*"

.PHONY:
release-changelog:
	github_changelog_generator -u fastly -p cli \
		--no-pr-wo-labels \
		--no-author \
		--no-issues \
		--enhancement-label "**Enhancements:**" \
		--bugs-label "**Bug fixes:**" \
		--release-url "https://github.com/fastly/cli/releases/tag/%s" \
		--exclude-labels documentation \
		--output RELEASE_CHANGELOG.md \
		--since-tag $(PREVIOUS_SEMVER_TAG)
