export GO111MODULE:=on
export CGO_ENABLED:=0

BIN      = bin
DIST     = dist
GO       = go
GORELEASER = $(BIN)/goreleaser

.PHONY: all
all: fmt tidy build test

.PHONY: build
build:
	@go build -o cola ./cmd/cola

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
	@rm -rf dist

$(BIN):
	@mkdir -p $@

$(BIN)/%: | $(BIN)
	env GOBIN=$(abspath $(BIN)) $(GO) install $(PACKAGE)

$(BIN)/goreleaser: PACKAGE=github.com/goreleaser/goreleaser/v2@latest

.PHONY: release
release: $(GORELEASER)
	$(GORELEASER) release --clean

.PHONY: snapshot
snapshot: $(GORELEASER)
	$(GORELEASER) release --snapshot --clean

.PHONY: docs
docs:
	@mkdir -p dist/docs
	@docker build -t cola-docs-builder:local ./docs/
	@docker run -v $(PWD):/documents/ cola-docs-builder:local asciidoctor-multipage -r asciidoctor-multipage -D dist/docs/ docs/index.adoc
