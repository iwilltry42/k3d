SHELL := /bin/bash

# Build targets
TARGETS ?= darwin/amd64 linux/amd64 linux/386 linux/arm linux/arm64 windows/amd64
TARGET_OBJS ?= darwin-amd64.tar.gz darwin-amd64.tar.gz.sha256 linux-amd64.tar.gz linux-amd64.tar.gz.sha256 linux-386.tar.gz linux-386.tar.gz.sha256 linux-arm.tar.gz linux-arm.tar.gz.sha256 linux-arm64.tar.gz linux-arm64.tar.gz.sha256 windows-amd64.zip windows-amd64.zip.sha256

# get git tag
GIT_TAG   := $(shell git describe --tags)
ifeq ($(GIT_TAG),)
GIT_TAG   := $(shell git describe --always)
endif

# get latest k3s version: grep the tag JSON field, extract the tag and replace + with - (difference between git and dockerhub tags)
K3S_TAG		:= $(shell curl --silent "https://api.github.com/repos/rancher/k3s/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed -E 's/\+/\-/')
ifeq ($(K3S_TAG),)
$(warning K3S_TAG undefined: couldn't get latest k3s image tag!)
$(warning Output of curl: $(shell curl --silent "https://api.github.com/repos/rancher/k3s/releases/latest"))
$(error exiting)
endif

# determine if make is being executed from interactive terminal
INTERACTIVE:=$(shell [ -t 0 ] && echo 1)

# Go options
GO        ?= go
PKG       := $(shell go mod vendor)
TAGS      :=
TESTS     := .
TESTFLAGS :=
LDFLAGS   := -w -s -X github.com/rancher/k3d/version.Version=${GIT_TAG} -X github.com/rancher/k3d/version.K3sVersion=${K3S_TAG}
GOFLAGS   :=
BINDIR    := $(CURDIR)/bin
BINARIES  := k3d


# Go Package required
PKG_GOX := github.com/mitchellh/gox@v1.0.1
PKG_GOLANGCI_LINT_VERSION := 1.23.1
PKG_GOLANGCI_LINT := github.com/golangci/golangci-lint/cmd/golangci-lint@v${PKG_GOLANGCI_LINT_VERSION}

# configuration adjustments for golangci-lint
GOLANGCI_LINT_DISABLED_LINTERS := "" # disabling typecheck, because it currently (06.09.2019) fails with Go 1.13

# Use Go Modules for everything
export GO111MODULE=on

# go source directories.
# DIRS defines a single level directly, we only look at *.go in this directory.
# REC_DIRS defines a source code tree. All go files are analyzed recursively.
DIRS :=  .
REC_DIRS := cmd pkg

# Rules for finding all go source files using 'DIRS' and 'REC_DIRS'
GO_SRC := $(foreach dir,$(DIRS),$(wildcard $(dir)/*.go))
GO_SRC += $(foreach dir,$(REC_DIRS),$(shell find $(dir) -name "*.go"))

# Rules for directory list as input for the golangci-lint program
LINT_DIRS := $(DIRS) $(foreach dir,$(REC_DIRS),$(dir)/...)

.PHONY: all build build-cross clean fmt check-fmt lint check extra-clean install-tools

all: clean fmt check build

build:
	CGO_ENABLED=0 $(GO) build -i $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o '$(BINDIR)/$(BINARIES)'

build-cross: LDFLAGS += -extldflags "-static"
build-cross:
	CGO_ENABLED=0 gox -parallel=3 -output="_dist/$(BINARIES)-{{.OS}}-{{.Arch}}" -osarch='$(TARGETS)' $(GOFLAGS) $(if $(TAGS),-tags '$(TAGS)',) -ldflags '$(LDFLAGS)'

clean:
	@rm -rf $(BINDIR) _dist/

extra-clean: clean
	$(GO) clean -i $(PKG_GOX)
	$(GO) clean -i $(PKG_GOLANGCI_LINT)

# fmt will fix the golang source style in place.
fmt:
	@gofmt -s -l -w $(GO_SRC)

# check-fmt returns an error code if any source code contains format error.
check-fmt:
	@test -z $(shell gofmt -s -l $(GO_SRC) | tee /dev/stderr) || echo "[WARN] Fix formatting issues with 'make fmt'"

lint:
	@golangci-lint run -D $(GOLANGCI_LINT_DISABLED_LINTERS) $(LINT_DIRS)

check: check-fmt lint

# Check for required executables
HAS_GOX := $(shell command -v gox 2> /dev/null)
HAS_GOLANGCI  := $(shell command -v golangci-lint)
HAS_GOLANGCI_VERSION := $(shell golangci-lint --version | grep "version $(PKG_GOLANGCI_LINT_VERSION) " 2>&1)

install-tools:
ifndef HAS_GOX
	($(GO) get $(PKG_GOX))
endif
ifndef HAS_GOLANGCI
	(curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b ${GOPATH}/bin v${PKG_GOLANGCI_LINT_VERSION})
endif
ifdef HAS_GOLANGCI
ifeq ($(HAS_GOLANGCI_VERSION),)
ifdef INTERACTIVE
	@echo "Warning: Your installed version of golangci-lint (interactive: ${INTERACTIVE}) differs from what we'd like to use. Switch to v${PKG_GOLANGCI_LINT_VERSION}? [Y/n]"
	@read line; if [ $$line == "y" ]; then (curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b ${GOPATH}/bin v${PKG_GOLANGCI_LINT_VERSION}); fi
else
	@echo "Warning: you're not using the same version of golangci-lint as us (v${PKG_GOLANGCI_LINT_VERSION})"
endif
endif
endif
