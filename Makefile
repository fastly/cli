.PHONY: clean

SHELL := /bin/bash -o pipefail ## Set the shell to use for finding Go files (default: /bin/bash)

# Compile program (implicit default target).
#
# GO_ARGS allows for passing additional arguments.
# e.g. make build GO_ARGS='--ldflags "-s -w"'
.PHONY: build
build: config ## Compile program (CGO disabled)
	CGO_ENABLED=0 $(GO_BIN) build $(GO_ARGS) ./cmd/fastly

GO_BIN ?= go ## Allows overriding go executable.
TEST_COMMAND ?= $(GO_BIN) test ## Enables support for tools such as https://github.com/rakyll/gotest
TEST_ARGS ?= -timeout 15m ./... ## The compute tests can sometimes exceed the default 10m limit

GOHOSTOS ?= $(shell go env GOHOSTOS || echo unknown)
GOHOSTARCH ?= $(shell go env GOHOSTARCH || echo unknown)

ifeq ($(OS), Windows_NT)
	SHELL = cmd.exe
	.SHELLFLAGS = /c
	GO_FILES = $(shell where /r pkg *.go)
	GO_FILES += $(shell where /r cmd *.go)
	CONFIG_SCRIPT = scripts\config.sh
	CONFIG_FILE = pkg\config\config.toml
else
	GO_FILES = $(shell find cmd pkg -type f -name '*.go')
	CONFIG_SCRIPT = ./scripts/config.sh
	CONFIG_FILE = pkg/config/config.toml
endif

# Build executables using goreleaser (useful for local testing purposes).
#
# You can pass flags to goreleaser via GORELEASER_ARGS
# --clean will save you deleting the dist dir
# --single-target will be quicker and only build for your os & architecture
# --skip-post-hooks which prevents errors such as trying to execute the binary for each OS (e.g. we call scripts/documentation.sh and we can't run Windows exe on a Mac).
# --skip-validate will skip the checks (e.g. git tag checks which result in a 'dirty git state' error)
#
# EXAMPLE:
# make release GORELEASER_ARGS="--clean --skip-post-hooks --skip-validate"
release: dependencies $(GO_FILES) ## Build executables using goreleaser
	@GOHOSTOS="${GOHOSTOS}" GOHOSTARCH="${GOHOSTARCH}" goreleaser build ${GORELEASER_ARGS}

# Useful for attaching a debugger such as https://github.com/go-delve/delve
debug:
	@$(GO_BIN) build -gcflags="all=-N -l" $(GO_ARGS) -o "fastly" ./cmd/fastly

.PHONY: config
config:
	@$(CONFIG_SCRIPT)

.PHONY: all
all: config dependencies tidy fmt vet staticcheck gosec semgrep test build install ## Run EVERYTHING!

# Update CI tools used by ./.github/workflows/pr_test.yml
.PHONY: dependencies
dependencies:
	@while read -r line || [ -n "$$line" ]; do \
		$(GO_BIN) install $$line; \
	done < .github/dependencies.txt
	@if [ "$$(uname)" = 'Darwin' ]; then \
			if ! command -v semgrep &> /dev/null; then \
					brew install semgrep; \
			fi \
	fi

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
vet: config ## Run vet static analysis
	$(GO_BIN) vet ./{cmd,pkg}/...

# Run linter.
.PHONY: revive
revive: ## Run linter (using revive)
	revive ./...

# Run security vulnerability checker.
.PHONY: gosec
gosec: ## Run security vulnerability checker
	gosec -quiet -exclude=G104 ./{cmd,pkg}/...

# Run semgrep checker.
# NOTE: We can only exclude the import-text-template rule via a semgrep CLI flag
.PHONY: semgrep
semgrep: ## Run semgrep
	@if [ '$(SEMGREP_SKIP)' != 'true' ]; then \
		if command -v semgrep &> /dev/null; then semgrep ci --config auto --exclude-rule go.lang.security.audit.xss.import-text-template.import-text-template $(SEMGREP_ARGS); fi \
	fi

# Run third-party static analysis.
# To ignore lines use: //lint:ignore <CODE> <REASON>
.PHONY: staticcheck
staticcheck: ## Run static analysis
	staticcheck ./{cmd,pkg}/...

# Run tests
.PHONY: test
test: config ## Run tests (with race detection)
	@$(TEST_COMMAND) -race $(TEST_ARGS)

# Compile and install program.
#
# GO_ARGS allows for passing additional arguments.
.PHONY: install
install: config ## Compile and install program
	CGO_ENABLED=0 $(GO_BIN) install $(GO_ARGS) ./cmd/fastly

# Scaffold a new CLI command from template files.
.PHONY: scaffold
scaffold:
	@$(shell pwd)/scripts/scaffold.sh $(CLI_PACKAGE) $(CLI_COMMAND) $(CLI_API)

# Scaffold a new CLI 'category' command from template files.
.PHONY: scaffold-category
scaffold-category:
	@$(shell pwd)/scripts/scaffold-category.sh $(CLI_CATEGORY) $(CLI_CATEGORY_COMMAND) $(CLI_PACKAGE) $(CLI_COMMAND) $(CLI_API)

# Graph generates a call graph that focuses on the specified package.
# Output is callvis.svg
# e.g. make graph PKG_IMPORT_PATH=github.com/fastly/cli/pkg/commands/kvstoreentry
.PHONY: graph
graph: ## Graph generates a call graph that focuses on the specified package
	@$(GO_BIN) install github.com/ofabry/go-callvis@latest 2>/dev/null
	go-callvis -file "callvis" -focus "$(PKG_IMPORT_PATH)" ./cmd/fastly/
	@rm callvis.gv

.PHONY: help
help:
	@printf "Targets\n"
	@(grep -h -E '^[0-9a-zA-Z_.-]+:.*?## .*$$' $(MAKEFILE_LIST) || true) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'
	@printf "\nDefault target\n"
	@printf "\033[36m%s\033[0m" $(.DEFAULT_GOAL)
	@printf "\n\nMake Variables\n"
	@(grep -h -E '^[0-9a-zA-Z_.-]+\s[:?]?=.*? ## .*$$' $(MAKEFILE_LIST) || true) | sort | awk 'BEGIN {FS = "[:?]?=.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'
