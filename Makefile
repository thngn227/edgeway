BINARY  := edgeway
MODULE  := edgeway
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS := -s -w -X $(MODULE)/internal/version.Version=$(VERSION) -X $(MODULE)/internal/version.Commit=$(COMMIT)
COMPOSE := docker compose -f deploy/docker-compose.yml

.DEFAULT_GOAL := help
.PHONY: help build run test vet fmt fmt-check check tidy up down docker clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "} {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

build: ## Build bin/edgeway
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/edgeway

run: build ## Build and run with configs/edgeway.yaml
	./bin/$(BINARY) -config configs/edgeway.yaml

test: ## Run unit tests with the race detector
	go test -race ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format all Go files
	gofmt -w .

fmt-check: ## Fail if any Go file is not formatted
	@unformatted="$$(gofmt -l .)"; if [ -n "$$unformatted" ]; then echo "Run 'make fmt':"; echo "$$unformatted"; exit 1; fi

check: fmt-check vet test build ## Everything CI runs

tidy: ## Tidy go.mod / go.sum
	go mod tidy

up: ## Start dev backends (docker compose)
	$(COMPOSE) up -d

down: ## Stop dev backends
	$(COMPOSE) down

docker: ## Build the edgeway Docker image
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) -t $(BINARY):$(VERSION) .

clean: ## Remove build output
	rm -rf bin
