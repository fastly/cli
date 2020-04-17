SHELL := /bin/bash -o pipefail

LAST_SEMVER_TAG := $(shell git tag | sort -r --version-sort | egrep '^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$$' | head -n1)

.PHONY: all
all: fmt vet staticcheck lint gosec test build install

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
		--since-tag $(LAST_SEMVER_TAG)
