export CGO_ENABLED:=0
export GO111MODULE:=on

MODULE   = $(shell $(GO) list -m)
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo v0)
SCHEMAS := $(wildcard pkg/types/schemas/*.json)
# BINARIES := $(addprefix dist/,$(patsubst cmd/%,%,$(wildcard cmd/*)))

BIN      = bin
DIST     = dist
GO       = go

BINARIES := dist/cola-ignition

.PHONY: all
all: fmt tidy build test

.PHONY: build
build: $(BINARIES)

.PHONY: test
test:
	@go test ./... -cover

.PHONY: fmt
fmt:
	@test -z $$(go fmt ./...)

.PHONY: lint
lint:
	@golangci-lint run ./...

.PHONY: tidy
tidy:
	@go mod tidy

clean:
	@rm -f $(BINARIES)

.PHONY: generate
generate:
	@go-jsonschema \
		--struct-name-from-title \
		--extra-imports \
		-p github.com/tmacro/sysctr/pkg/types \
		-o pkg/types/types_gen.go \
		$(SCHEMAS)

$(DIST):
	@mkdir -p $@

$(DIST)/%: | $(DIST)
	$(GO) build \
		-tags release \
		-ldflags '-X $(MODULE)/$(patsubst dist/%,cmd/%,$@).Version=$(VERSION) -X $(MODULE)/cmd.BuildDate=$(DATE)' \
		-o $@ ./$(patsubst dist/%,cmd/%,$@)

$(BIN):
	@mkdir -p $@

$(BIN)/%: | $(BIN)
	env GOBIN=$(abspath $(BIN)) $(GO) install $(PACKAGE)

$(BIN)/goreleaser: PACKAGE=github.com/goreleaser/goreleaser/v2@latest

GORELEASER = $(BIN)/goreleaser

release: $(GORELEASER)
	$(GORELEASER) release --clean

snapshot: $(GORELEASER)
	$(GORELEASER) release --snapshot --clean
