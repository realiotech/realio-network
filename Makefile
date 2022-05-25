PACKAGES=$(shell go list ./...)
COVERAGE ?= coverage.txt
GOPATH ?= $(shell $(GO) env GOPATH)
GOBIN ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build
NETWORK ?= testnet
REALIO_NETWORK_BINARY = realio-networkd
TESTNET_FLAGS ?=
VERSION := $(shell echo $(shell git describe --tags 2>/dev/null ) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
HTTPS_GIT := https://github.com/realiotech/realio-network.git

export GO111MODULE = on

# Default target executed when no arguments are given to make.
default_target: all

.PHONY: default_target

# process build tags
build_tags = netgo
ifeq ($(NETWORK),mainnet)
    build_tags += mainnet
else ifeq ($(NETWORK),testnet)
    build_tags += testnet
endif


build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace := $(subst ,, )
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags
ldflags += -X github.com/cosmos/cosmos-sdk/version.Name=realionetwork \
	-X github.com/cosmos/cosmos-sdk/version.AppName=$(REALIO_NETWORK_BINARY) \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
	-X github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

###############################################################################
###                                  Build                                  ###
###############################################################################


all: build
build: check-network go.sum
	@go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/realio-networkd ./cmd/realio-networkd

install: check-network go.sum
	@go install -mod=readonly $(BUILD_FLAGS) ./cmd/realio-networkd

test:
	@go test -v -mod=readonly $(PACKAGES) -coverprofile=$(COVERAGE) -covermode=atomic

.PHONY: clean build install test

clean:
	@echo "Cleaning up old build and daemons"
	rm -rf $(BUILDDIR)/
	$(RM) $(GOBIN)/realio-networkd


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

DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf

containerProtoVer=v0.3
containerProtoImage=tendermintdev/sdk-proto-gen:$(containerProtoVer)
containerProtoGen=cosmos-sdk-proto-gen-$(containerProtoVer)
containerProtoGenSwagger=cosmos-sdk-proto-gen-swagger-$(containerProtoVer)
containerProtoFmt=cosmos-sdk-proto-fmt-$(containerProtoVer)

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) sh ./scripts/protocgen.sh

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) sh ./scripts/protoc-swagger-gen.sh

proto-format:
	@echo "Formatting Protobuf files"
	find ./ -not -path "./third_party/*" -name *.proto -exec clang-format -i {} \;

proto-lint:
	@$(DOCKER_BUF) lint --error-format=json

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against $(HTTPS_GIT)#branch=main


TM_URL              = https://raw.githubusercontent.com/tendermint/tendermint/v0.34.12/proto/tendermint
GOGO_PROTO_URL      = https://raw.githubusercontent.com/regen-network/protobuf/cosmos
COSMOS_SDK_URL      = https://raw.githubusercontent.com/cosmos/cosmos-sdk/v0.44.0
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
## Issue link: https://github.com/tendermint/tendermint/issues/5021
	@mkdir -p $(TM_ABCI_TYPES)
	@curl -sSL $(TM_URL)/abci/types.proto > $(TM_ABCI_TYPES)/types.proto

	@mkdir -p $(TM_TYPES)
	@curl -sSL $(TM_URL)/types/types.proto > $(TM_TYPES)/types.proto

	@mkdir -p $(TM_CRYPTO_TYPES)
	@curl -sSL $(TM_URL)/crypto/proof.proto > $(TM_CRYPTO_TYPES)/proof.proto
	@curl -sSL $(TM_URL)/crypto/keys.proto > $(TM_CRYPTO_TYPES)/keys.proto

.PHONY: proto-all proto-gen proto-format proto-lint proto-check-breaking proto-update-deps