#! /usr/bin/env make
.PHONY: help fmt lint vet

VERSION := $(shell git describe --tags --long --always --match "v*.*.*")
API_URL ?= "https://api.cloud-guardian.net/cloudguardian-api/v1/"
LDFLAGS := -X 'patchmaster-client/cli.Version=$(VERSION)' -X 'patchmaster-client/cli.ApiUrl=$(API_URL)'
SRC_FILES = $(shell find . -type f -name '*.go')

help: ## Displays help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-z0-9A-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples build targets for docker:"
	@echo "make run-docker-almalinux-9.3"
	@echo "make run-docker-rockylinux-9.3"
	@echo "make run-docker-rockylinux-9.5"
	@echo "make run-docker-rockylinux-9.5.20241118"
	@echo "make run-docker-ubuntu-jammy-20240808"

tidy: ## Refresh go.mod file
	go mod tidy

clean: ## Cleanup environment
	@rm -f bin/*/*

fix: fmt

run: ## Run main.go
	go run main.go

fmt: ## Format code (runs "go fmt")
	gofmt -w .

lint: ## Run linting (runs golangci-lint run main.go)
	gofmt -l -d .

vet: ## Run vet (runs "go vet")
	go vet main.go

setup: tidy ## Install developer requirements

test: setup ## Recursively run go test
	go test ./...

test_verbose: setup ## Recursively run go test, with verbosity
	go test -v ./...

dev: setup ## Run application for local development

binary=cloud-guardian

build_archs = \
	linux_amd64 \
	linux_arm64 \
	darwin_amd64 \
	darwin_arm64

bin/% /bin%.exe: ${SRC_FILES} ## Build binary for the specified architecture
	$(eval OSARCH = $(subst /, ,$*))
	$(eval OSARCH = $(subst _, ,${OSARCH}))
	GOOS=$(word 1, $(OSARCH)) GOARCH=$(word 2, $(OSARCH)) go build -ldflags="${LDFLAGS}" -o $@ ${SRC_MAIN}

build: $(foreach arch, ${build_archs}, bin/${arch}/${binary})
build_windows: bin/windows_amd64/${binary}.exe

release: build ## Create builds for Linux

# Docker test environments
DOCKER_PLATFORMS := --platform linux/amd64
DOCKER_VOLUME := --volume ./bin:/client
DOCKER_COMMON_FLAGS := --rm -it $(DOCKER_PLATFORMS) $(DOCKER_VOLUME) --env API_KEY=${API_KEY}

run-docker-almalinux-%: ## Run AlmaLinux container
	docker run $(DOCKER_COMMON_FLAGS) docker.io/almalinux:$*

run-docker-rockylinux-%: ## Run Rocky Linux container
	docker run $(DOCKER_COMMON_FLAGS) docker.io/rockylinux/rockylinux:$*

run-docker-ubuntu-%: ## Run Ubuntu container
	docker run $(DOCKER_COMMON_FLAGS) docker.io/ubuntu:jammy-20240808
