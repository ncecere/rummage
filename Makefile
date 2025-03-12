# Makefile for Rummage

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOFMT=gofmt
GOLINT=golangci-lint
BINARY_NAME=rummage
BINARY_UNIX=$(BINARY_NAME)_unix
MAIN_PATH=./cmd/rummage

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build clean test coverage lint fmt vet tidy help

all: test build

build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

clean: ## Remove build artifacts
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f coverage.out

test: ## Run tests
	$(GOTEST) -v ./...

coverage: ## Run tests with coverage
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

lint: ## Run linter
	$(GOLINT) run

fmt: ## Format code
	$(GOFMT) -s -w .

vet: ## Run go vet
	$(GOVET) ./...

tidy: ## Tidy go.mod
	$(GOMOD) tidy

run: ## Run the application
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)
	./$(BINARY_NAME)

docker-build: ## Build docker image
	docker build -t $(BINARY_NAME) .

docker-run: ## Run docker container
	docker run -p 8080:8080 $(BINARY_NAME)

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help
