GO ?= $(shell command -v go 2> /dev/null)
GO_TEST_FLAGS ?= -race
GO_BUILD_FLAGS ?=
IMAGE_NAME ?= server:latest

.PHONY: all
all: check-style test build

.PHONY: lint-install
lint-install:
	@echo Installing golangci-lint
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2

.PHONY: check-style
check-style:
	@echo Running golangci-lint
	@golangci-lint run ./...

.PHONY: lint-fix
lint-fix:
	@echo Running golangci-lint with auto-fix
	@golangci-lint run --fix ./...

.PHONY: test
test:
	@echo Running tests
	@$(GO) test -v $(GO_TEST_FLAGS) ./...

.PHONY: build
build:
	@echo Building for linux amd64
	@env GOOS=linux GOARCH=amd64 $(GO) build -trimpath -o build/dist/server-linux-amd64 ./cmd/server;
	@echo Building for linux arm64
	@env GOOS=linux GOARCH=arm64 $(GO) build -trimpath -o build/dist/server-linux-arm64 ./cmd/server;
	@echo Building for darwin amd64
	@env GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -o build/dist/server-darwin-amd64 ./cmd/server;
	@echo Building for darwin arm64
	@env GOOS=darwin GOARCH=arm64 $(GO) build -trimpath -o build/dist/server-darwin-arm64 ./cmd/server;
	@echo Building for windows amd64
	@env GOOS=windows GOARCH=amd64 $(GO) build -trimpath -o build/dist/server-windows-amd64.exe ./cmd/server;

.PHONY: run
run:
	air -c ./initializer/.air.toml

.PHONY: image
image:
	@echo Building a docker image
	@docker build -f deployments/Dockerfile -t $(IMAGE_NAME) .
