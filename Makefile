SHELL=/bin/bash -e -o pipefail
ARGS=$(filter-out $@,$(MAKECMDGOALS))
LDFLAGS=-ldflags "-s -w \
	-X main.Version=${BUILD_VERSION} \
	-X main.BuildDate=`date -u +%Y-%m-%d.%H:%M:%S` \
	-X main.GitBranch=${CI_COMMIT_BRANCH} \
	-X main.GitCommit=${CI_COMMIT_SHORT_SHA}"

.DEFAULT_GOAL=help

help: ## Show help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build binary
	CGO_ENABLED=0 go build ${LDFLAGS} ./...

test: ## Run unit tests
	@go test -failfast -race -v ./...

lint: ## Run linter
	@golangci-lint run --timeout 5m

genmock: ## Generate all mocks
	@mockery
