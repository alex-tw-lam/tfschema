SHELL := /bin/bash

.PHONY: test cover lint complexity vuln vet fmt tidy help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

test: ## Run the test suite with the race detector
	go test -race ./...

cover: ## Run tests with coverage and enforce a coverage floor
	go test -race -coverprofile=cover.out -coverpkg=./... ./...
	@go tool cover -func=cover.out | tail -1
	go-test-coverage --config=.testcoverage.yml

lint: ## Run golangci-lint (code quality + cyclomatic complexity + unused code)
	golangci-lint run ./...

complexity: ## Report functions with cyclomatic complexity over 15 (informational)
	@gocyclo -over 15 cmd internal tests 2>/dev/null || true

vuln: ## Scan dependencies for known vulnerabilities (govulncheck)
	govulncheck ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format Go source
	gofmt -s -w .

tidy: ## Tidy modules and refresh the vendor directory
	go mod tidy
	go mod vendor
