# SPDX-License-Identifier: Apache-2.0
# Copyright 2019 Open Networking Foundation
# Copyright 2024 Intel Corporation

export CGO_ENABLED=1
export GO111MODULE=on

.PHONY: build

ONOS_PCI_VERSION ?= latest
ONOS_PROTOC_VERSION := v0.6.6
BUF_VERSION := 0.27.1

GOLANG_CI_VERSION := v1.52.2

all: build docker-build

build: # @HELP build the Go binaries and run all validations (default)
	GOPRIVATE="github.com/onosproject/*" go build -o build/_output/onos-pci ./cmd/onos-pci

test: # @HELP run the unit tests and source code validation
test: build lint license
	go test -race github.com/onosproject/onos-pci/pkg/...
	go test -race github.com/onosproject/onos-pci/cmd/...

docker-build-onos-pci: # @HELP build onos-pci Docker image
	@go mod vendor
	docker build . -f build/onos-pci/Dockerfile \
		-t onosproject/onos-pci:${ONOS_PCI_VERSION}
	@rm -rf vendor

docker-build: # @HELP build all Docker images
docker-build: build docker-build-onos-pci

docker-push-onos-pci: # @HELP push onos-pci Docker image
	docker push onosproject/onos-pci:${ONOS_PCI_VERSION}

docker-push: # @HELP push docker images
docker-push: docker-push-onos-pci

lint: # @HELP examines Go source code and reports coding problems
	golangci-lint --version | grep $(GOLANG_CI_VERSION) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b `go env GOPATH`/bin $(GOLANG_CI_VERSION)
	golangci-lint run --timeout 15m

license: # @HELP run license checks
	rm -rf venv
	python3 -m venv venv
	. ./venv/bin/activate;\
	python3 -m pip install --upgrade pip;\
	python3 -m pip install reuse;\
	reuse lint

check-version: # @HELP check version is duplicated
	./build/bin/version_check.sh all

clean: # @HELP remove all the build artifacts
	rm -rf ./build/_output ./vendor ./cmd/onos-pci/onos-pci ./cmd/onos/onos venv
	go clean github.com/onosproject/onos-pci/...

help:
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST) \
    | sort \
    | awk ' \
        BEGIN {FS = ": *# *@HELP"}; \
        {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}; \
    '