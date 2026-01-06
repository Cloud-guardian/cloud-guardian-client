#! /usr/bin/env make
.PHONY: help fmt lint vet

VERSION ?= $(shell git describe --tags --long --always --match "*.*.*")
API_URL ?= "https://api.cloud-guardian.net/cloudguardian-api/v1/"
LDFLAGS := -X 'cloud-guardian/cloudguardian_version.Version=v$(VERSION)' -X 'cloud-guardian/cli.ApiUrl=$(API_URL)'
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
	@rm -f dist/*

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

releases = \
	${binary}_${VERSION}_linux_amd64.tar.gz \
	${binary}_${VERSION}_linux_arm64.tar.gz \
	${binary}_${VERSION}_darwin_amd64.zip \
	${binary}_${VERSION}_darwin_arm64.zip \
	${binary}_${VERSION}_windows_amd64.zip \
	${binary}_${VERSION}_windows_arm64.zip

build: $(foreach arch, ${build_archs}, bin/${arch}/${binary}) build_windows
build_windows: bin/windows_amd64/${binary}.exe bin/windows_arm64/${binary}.exe

bin/%: ${SRC_FILES} ## Build binary for the specified architecture
	$(eval OSARCH = $(subst /, ,$*))
	$(eval OSARCH = $(subst _, ,${OSARCH}))
	GOOS=$(word 1, $(OSARCH)) GOARCH=$(word 2, $(OSARCH)) go build -ldflags="${LDFLAGS}" -o $@ ${SRC_MAIN}

dist:
	@mkdir -p dist

release: build dist $(foreach release, ${releases}, dist/${release}) ## Create release archives in dist/

dist/%: ## Create zip / tarball archives from the binaries
	@echo "$@:"
	$(eval PARTS = $(subst _, ,$*))
	$(eval OS = $(word 3, $(PARTS)))
	$(eval PARTS = $(subst ., ,$(word 4, $(PARTS))))
	$(eval ARCH = $(word 1, $(PARTS)))
	$(eval EXT = $(suffix $(suffix $*)))
	@if [ "${EXT}" = ".zip" ]; then \
		if [ "${OS}" = "windows" ]; then \
			cd bin/${OS}_${ARCH} && zip ../../$@ ${binary}.exe; \
		else \
			cd bin/${OS}_${ARCH} && zip ../../$@ ${binary}; \
		fi; \
	else \
		tar --create --verbose --gzip --file $@ --directory bin/${OS}_${ARCH} ${binary}; \
	fi

# Docker test environments
DOCKER_PLATFORMS := --platform linux/amd64
DOCKER_VOLUME := --volume ./bin:/client
DOCKER_ENV := --env API_KEY=${API_KEY}
DOCKER_COMMON_FLAGS := --rm -it $(DOCKER_PLATFORMS) $(DOCKER_VOLUME) $(DOCKER_ENV)

run-docker-almalinux-%: ## Run AlmaLinux container
	docker run $(DOCKER_COMMON_FLAGS) docker.io/almalinux:$*

run-docker-rockylinux-%: ## Run Rocky Linux container
	docker run $(DOCKER_COMMON_FLAGS) docker.io/rockylinux/rockylinux:$*

run-docker-ubuntu-%: ## Run Ubuntu container
	docker run $(DOCKER_COMMON_FLAGS) docker.io/ubuntu:$*


CONTAINER_PORTS = \
	5001:ubuntu_18.04 \
	5002:ubuntu_20.04 \
	5003:ubuntu_22.04 \
	5004:ubuntu_25.04 \
	5005:ubuntu_25.10 \
	5006:rockylinux_8 \
	5007:rockylinux_9 \
	5008:almalinux_9 \

test_containers/.run.%: test_containers/.built.% ## Run the test container
	$(eval PORT_IMAGE := $(filter %:$*,$(CONTAINER_PORTS)))
	$(eval PORT := $(word 1,$(subst :, ,$(PORT_IMAGE))))
	podman run --platform linux/arm64 --cap-add AUDIT_WRITE --name $* --publish $(PORT):22 --detach --rm $(DOCKER_VOLUME) $(DOCKER_ENV) $*
	@touch $@

test_containers/.built.%:
	@docker build --platform linux/arm64 --tag $* --file test_containers/$*/Dockerfile
	@touch $@
