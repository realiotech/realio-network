PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
VERSION ?= $(shell echo $(shell git describe --tags --always) | sed 's/^v//')
TMVERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::')
EVM_DENOM := ario
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
BINDIR ?= $(GOPATH)/bin
REALIO_BINARY = realio-networkd
REALIO_DIR = realio-network
BUILDDIR ?= $(CURDIR)/build
HTTPS_GIT := https://github.com/realiotech/realio-network.git
DOCKER := $(shell which docker)
NAMESPACE := realiotech
PROJECT := realio-network
DOCKER := $(shell which docker)
DOCKER_IMAGE := $(NAMESPACE)/$(PROJECT)
COMMIT_HASH := $(shell git rev-parse --short=7 HEAD)
DOCKER_TAG := $(COMMIT_HASH)
# e2e env
MOUNT_PATH := $(shell pwd)/build/:/root/
E2E_SKIP_CLEANUP := false

export GO111MODULE = on

# Default target executed when no arguments are given to make.
default_target: all

.PHONY: default_target

# process build tags
build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=realio-network \
          -X github.com/cosmos/cosmos-sdk/version.AppName=$(REALIO_BINARY) \
          -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
          -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
          -X github.com/cometbft/cometbft/version.TMCoreSemVer=$(TMVERSION) \
		  -X github.com/evmos/os/x/evm/types.DefaultEVMDenom=$(EVM_DENOM)

# DB backend selection
ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq (badgerdb,$(findstring badgerdb,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=badgerdb
endif
# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  build_tags += rocksdb
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb
endif
# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += boltdb
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=boltdb
endif

# add build tags to linker flags
whitespace := $(subst ,, )
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))
ldflags += -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

# check if no optimization option is passed
# used for remote debugging
ifneq (,$(findstring nooptimization,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -gcflags "all=-N -l"
endif


###############################################################################
###                                  Build                                  ###
###############################################################################

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/

build-linux:
	GOOS=linux GOARCH=amd64 LEDGER_ENABLED=false $(MAKE) build

build-darwin:
	GOOS=darwin GOARCH=arm64 LEDGER_ENABLED=false $(MAKE) build

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	go $@ $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

distclean: clean tools-clean

clean:
	rm -rf \
    $(BUILDDIR)/ \
    artifacts/ \
    tmp-swagger-gen/

all: build

build-all: tools build lint test vulncheck

.PHONY: distclean clean build-all build


###############################################################################
###                                Localnet                                 ###
###############################################################################

# Build image for a local testnet
localnet-build:
	@$(MAKE) -C networks/local

# Start a 4-node testnet locally
localnet-start: localnet-stop localnet-build
	@if ! [ -f build/node0/$(REALIO_DIR)/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/$(REALIO_DIR):Z realio-network/node "./realio-networkd testnet init-files --v 4 -o /realio-network --keyring-backend=test --starting-ip-address 192.167.10.2"; fi
	docker-compose up -d

# Stop testnet
localnet-stop:
	docker-compose down

# Clean testnet
localnet-clean:
	docker-compose down
	sudo rm -rf build/*

 # Reset testnet
localnet-unsafe-reset:
	docker-compose down
ifeq ($(OS),Windows_NT)
	@docker run --rm -v $(CURDIR)\build\node0\realio-network:/realio-network\Z realio-network/node "./realio-networkd tendermint unsafe-reset-all --home=/realio-network"
	@docker run --rm -v $(CURDIR)\build\node1\realio-network:/realio-network\Z realio-network/node "./realio-networkd tendermint unsafe-reset-all --home=/realio-network"
	@docker run --rm -v $(CURDIR)\build\node2\realio-network:/realio-network\Z realio-network/node "./realio-networkd tendermint unsafe-reset-all --home=/realio-network"
	@docker run --rm -v $(CURDIR)\build\node3\realio-network:/realio-network\Z realio-network/node "./realio-networkd tendermint unsafe-reset-all --home=/realio-network"
else
	@docker run --rm -v $(CURDIR)/build/node0/realio-network:/realio-network:Z realio-network/node "./realio-networkd tendermint unsafe-reset-all --home=/realio-network"
	@docker run --rm -v $(CURDIR)/build/node1/realio-network:/realio-network:Z realio-network/node "./realio-networkd tendermint unsafe-reset-all --home=/realio-network"
	@docker run --rm -v $(CURDIR)/build/node2/realio-network:/realio-network:Z realio-network/node "./realio-networkd tendermint unsafe-reset-all --home=/realio-network"
	@docker run --rm -v $(CURDIR)/build/node3/realio-network:/realio-network:Z realio-network/node "./realio-networkd tendermint unsafe-reset-all --home=/realio-network"
endif

# Clean testnet
localnet-show-logstream:
	docker-compose logs --tail=1000 -f

.PHONY: localnet-build localnet-start localnet-stop

###############################################################################
###                                Releasing                                ###
###############################################################################

PACKAGE_NAME:=github.com/realiotech/realio-network
GOLANG_CROSS_VERSION  = v1.22
GOPATH ?= '$(HOME)/go'
release-dry-run:
	@mkdir -p /tmp/go-cache
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v /tmp/go-cache:/go/pkg \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		--clean --skip=validate --skip=publish --snapshot
	@rm -rf /tmp/go-cache

release:
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m";\
		exit 1;\
	fi
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		release --clean --skip=validate

.PHONY: release-dry-run release

###############################################################################
###                                Utility                                  ###
###############################################################################
check-network:
ifeq ($(NETWORK),mainnet)
else ifeq ($(NETWORK),testnet)
else
	@echo "Unrecognized network: ${NETWORK}"
endif
	@echo "building network: ${NETWORK}"


###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.14.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace --user 0 $(protoImageName)

# ------
# NOTE: If you are experiencing problems running these commands, try deleting
#       the docker images and execute the desired command again.
#

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	$(protoImage) sh ./scripts/protocgen.sh

proto-swagger-gen:
	@echo "Downloading Protobuf dependencies"
#	@make proto-download-deps
	@echo "Generating Protobuf Swagger"
	$(protoImage) sh ./scripts/protoc-swagger-gen.sh

proto-format:
	@echo "Formatting Protobuf files"
	$(protoImage) find ./ -name *.proto -exec clang-format -i {} \;

# NOTE: The linter configuration lives in .protolint.yaml
proto-lint:
	@echo "Linting Protobuf files"
	$(protoImage) lint ./proto

proto-check-breaking:
	@echo "Checking Protobuf files for breaking changes"
	$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main


TM_URL              = https://raw.githubusercontent.com/cometbft/cometbft/v0.38.11/proto/tendermint
GOGO_PROTO_URL      = https://raw.githubusercontent.com/regen-network/protobuf/cosmos
COSMOS_SDK_URL      = https://raw.githubusercontent.com/cosmos/cosmos-sdk/v0.50.9
COSMOS_PROTO_URL    = https://raw.githubusercontent.com/regen-network/cosmos-proto/master

TM_CRYPTO_TYPES     = third_party/proto/tendermint/crypto
TM_ABCI_TYPES       = third_party/proto/tendermint/abci
TM_TYPES            = third_party/proto/tendermint/types

GOGO_PROTO_TYPES    = third_party/proto/gogoproto

COSMOS_PROTO_TYPES  = third_party/proto/cosmos_proto

proto-update-deps:
	@mkdir -p $(GOGO_PROTO_TYPES)
	@curl -sSL $(GOGO_PROTO_URL)/gogoproto/gogo.proto > $(GOGO_PROTO_TYPES)/gogo.proto

	@mkdir -p $(COSMOS_PROTO_TYPES)
	@curl -sSL $(COSMOS_PROTO_URL)/cosmos.proto > $(COSMOS_PROTO_TYPES)/cosmos.proto

## Importing of tendermint protobuf definitions currently requires the
## use of `sed` in order to build properly with cosmos-sdk's proto file layout
## (which is the standard Buf.build FILE_LAYOUT)
## Issue link: https://github.com/cometbft/cometbft/issues/5021
	@mkdir -p $(TM_ABCI_TYPES)
	@curl -sSL $(TM_URL)/abci/types.proto > $(TM_ABCI_TYPES)/types.proto

	@mkdir -p $(TM_TYPES)
	@curl -sSL $(TM_URL)/types/types.proto > $(TM_TYPES)/types.proto

	@mkdir -p $(TM_CRYPTO_TYPES)
	@curl -sSL $(TM_URL)/crypto/proof.proto > $(TM_CRYPTO_TYPES)/proof.proto
	@curl -sSL $(TM_URL)/crypto/keys.proto > $(TM_CRYPTO_TYPES)/keys.proto

.PHONY: proto-all proto-gen proto-format proto-lint proto-check-breaking proto-update-deps