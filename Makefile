# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLINT=golangci-lint

# Targets
.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "Please use 'make <target>' where <target> is one of"
	@echo "  build      to build the Go project"
	@echo "  test       to run tests on the Go project"
	@echo "  lint       to run linters on the Go project"

.PHONY: build
build:
	$(GOBUILD) -o bin/yourapp ./cmd/yourapp

.PHONY: test
test:
	$(GOTEST) -v ./...

.PHONY: lint
lint:
	$(GOLINT) run ./...

.PHONY: install-tools
install-tools:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: all
all: lint test build

