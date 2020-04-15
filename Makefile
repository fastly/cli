SHELL := /bin/bash -o pipefail

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
