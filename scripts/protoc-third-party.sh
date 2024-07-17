#!/bin/bash
set -eo pipefail

# Define the package you are looking for
COSMOS_SDK_PACKAGE="github.com/cosmos/cosmos-sdk"
EVMOS_PACKAGE="github.com/evmos/evmos/v18"
IBC_PACKAGE="github.com/cosmos/ibc-go/v7"
MULTI_STAKING_PACKAGE="github.com/realio-tech/multi-staking-module"

# Function to get the replacement path from go.mod
get_replacement_path() {
  local package=$1
  local go_mod_file=$2

  # Search for the replacement directive in the go.mod file
  REPLACE_MODULE_PATH=$(grep -E "replace[[:space:]]+$package[[:space:]]+=>[[:space:]]+" "$go_mod_file" | awk '{print $4}')
  if [ -n "$REPLACE_MODULE_PATH" ]; then
    echo "$REPLACE_MODULE_PATH"
  else
    get_module_path "$package"
  fi
}

# Function to get the module path using go list
get_module_path() {
  local package=$1
  go list -m -f "{{.Dir}}" "$package"
}

# Path to the go.mod file (adjust this if go.mod is not in the current directory)
GO_MOD_FILE="./go.mod"

# Check if the package has a replacement
COSMOS_SDK_PATH=$(get_replacement_path "$COSMOS_SDK_PACKAGE" "$GO_MOD_FILE")
EVMOS_PATH=$(get_replacement_path "$EVMOS_PACKAGE" "$GO_MOD_FILE")
MULTI_STAKING_PATH=$(get_replacement_path "$MULTI_STAKING_PACKAGE" "$GO_MOD_FILE")
IBC_PATH=$(get_replacement_path "$IBC_PACKAGE" "$GO_MOD_FILE")

CURDIR=$(pwd)
mkdir third_party
cd third_party
mkdir proto

cp -r "$COSMOS_SDK_PATH/proto/cosmos" proto
cp -r "$COSMOS_SDK_PATH/proto/amino" proto
cp -r "$COSMOS_SDK_PATH/proto/tendermint" proto

cp -r "$EVMOS_PATH/proto/evmos" proto

cp -r "$EVMOS_PATH/proto/ethermint" proto

cp -r "$IBC_PATH/proto/ibc" proto

cp -r "$MULTI_STAKING_PATH/proto/multistaking" proto

cp "$CURDIR/proto/buf.third.party.yaml" proto/buf.yaml
cp "$CURDIR/proto/buf.gen.gogo.yaml" proto
cp "$CURDIR/proto/buf.gen.swagger.yaml" proto

find proto/ -type d -print0 | xargs -0 chmod 0755
find proto/cosmos -type f -print0 | xargs -0 chmod 0644

rm -rf proto/cosmos/mint

# fix missing cosmos/ics23/v1/proofs.proto file
cd proto/cosmos/
mkdir ics23
cd ics23
mkdir v1
cd v1
wget "https://raw.githubusercontent.com/cosmos/ics23/master/proto/cosmos/ics23/v1/proofs.proto"
cd "$CURDIR/third_party/proto"