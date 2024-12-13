export GO111MODULE:=on

BIN      = bin
DIST     = dist
GO       = go
GORELEASER = $(BIN)/goreleaser

.PHONY: all
all: fmt tidy build test

.PHONY: build
build: snapshot

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
