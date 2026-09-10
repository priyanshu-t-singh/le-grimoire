SHELL := /bin/bash

BINARY_NAME=le-grimoire
BUILD_DIR=bin
RELEASE_DIR=release
MAIN_FILE=main.go
DOCKER_USERNAME=priyanshu9943
DOCKER_IMAGE=le-grimoire

DOCKER_PLATFORM ?= linux/amd64,linux/arm64
VERSION         ?= dev
OUTPUT_TAR      ?= $(DOCKER_IMAGE)-$(VERSION).tar
FULL_IMAGE_NAME := $(DOCKER_USERNAME)/$(DOCKER_IMAGE)

LDFLAGS := -w -s -X 'le-grimoire/internal/constants.Version=$(VERSION)'

# $(1)=GOOS $(2)=GOARCH $(3)=output suffix (e.g. .exe or empty)
define cross_build
	@echo "Building for $(1)/$(2)..."
	@mkdir -p $(RELEASE_DIR)/$(VERSION)
	GOOS=$(1) GOARCH=$(2) go build \
		-ldflags="$(LDFLAGS)" \
		-o $(RELEASE_DIR)/$(VERSION)/$(BINARY_NAME)-$(VERSION)-$(1)-$(2)$(3) $(MAIN_FILE)
endef

.DEFAULT_GOAL := build

.PHONY: help build run clean clean-all dev test docker-build docker-run docker-save \
	build-linux build-windows build-mac build-all checksums register-device fmt vet lint

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Run the app without building a binary
	go run $(MAIN_FILE)

build: ## Build the binary for the current platform
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

run: build ## Build and run the binary
	./$(BUILD_DIR)/$(BINARY_NAME)

test: ## Run tests
	go test ./... -v

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)

clean-all: clean ## Remove build and release artifacts
	rm -rf $(RELEASE_DIR)

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Vet code
	@echo "Vetting code..."
	go vet ./...

lint: ## Run golangci-lint (requires installation)
	if ! command -v golangci-lint &> /dev/null; then \
		echo "golangci-lint not found. Please install it first."; \
		echo "You can install it by running: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	@echo "Linting..."
	golangci-lint run

docker-build: ## Build and push a multi-platform Docker image
	@echo "Building Docker image..."
	docker buildx build \
		--platform $(DOCKER_PLATFORM) \
		-t $(FULL_IMAGE_NAME):$(VERSION) \
		-t $(FULL_IMAGE_NAME):latest \
		--push .

docker-run: ## Run the Docker container for the current VERSION
	@echo "Running Docker container..."
	docker run \
		-p 8080:8080 \
		--name $(BINARY_NAME) \
		--rm \
		-v $(PWD)/.db:/app/.db \
		$(FULL_IMAGE_NAME):$(VERSION)

docker-save: ## Save the Docker image to a tarball
	@echo "Saving $(FULL_IMAGE_NAME):$(VERSION) to $(OUTPUT_TAR)..."
	docker save -o $(OUTPUT_TAR) $(FULL_IMAGE_NAME):$(VERSION)
	@echo "Saved successfully: $(OUTPUT_TAR)"

build-linux: ## Cross-compile for Linux (amd64, arm64)
	$(call cross_build,linux,amd64)
	$(call cross_build,linux,arm64)

build-windows: ## Cross-compile for Windows (amd64, arm64)
	$(call cross_build,windows,amd64,.exe)
	$(call cross_build,windows,arm64,.exe)

build-mac: ## Cross-compile for macOS (amd64, arm64)
	$(call cross_build,darwin,amd64)
	$(call cross_build,darwin,arm64)

build-all: build-linux build-windows build-mac checksums ## Cross-compile for all platforms and generate checksums

checksums: ## Generate SHA256 checksums for release binaries
	@echo "Generating checksums..."
	@cd $(RELEASE_DIR)/$(VERSION) && shasum -a 256 * > checksums.txt
	@echo "Checksums written to $(RELEASE_DIR)/$(VERSION)/checksums.txt"

register-device: build ## Register a device (usage: make register-device ID=... KEY=...)
	./$(BUILD_DIR)/$(BINARY_NAME) register-device --id "$(ID)" --key "$(KEY)"
