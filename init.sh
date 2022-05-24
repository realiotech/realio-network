KEY="mykey"
CHAINID="realionetwork_9000-1"
MONIKER="realionetworklocal"
KEYRING="test"
HOMEDIR="~/.realio-network"
KEYALGO="eth_secp256k1"
LOGLEVEL="trace"
# to trace evm
#TRACE="--trace"
TRACE=""

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

# remove remove existing node
rm -rf $HOME/.realio-network

# Reinstall daemon
#make clean
#make install

# Set moniker and chain-id for Realio Network (Moniker can be anything, chain-id must be an integer)
realio-networkd init $MONIKER --chain-id $CHAINID

# Set client config
realio-networkd config keyring-backend $KEYRING --home $HOMEDIR
realio-networkd config chain-id $CHAINID --home $HOMEDIR

# if $KEY exists it should be deleted
realio-networkd keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --home $HOMEDIR

# Change parameter token denominations to urio
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="urio,urst"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="urio"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="urio"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="urio"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="urio"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["inflation"]["params"]["mint_denom"]="urio"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json

# Add denom metadata for rio and rst
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["bank"]["denom_metadata"]=[{ "description": "The native token of the Realio Network", "denom_units": [ { "denom": "urio", "exponent": 0, "aliases": [ "microrio" ] }, { "denom": "rio", "exponent": 6, "aliases": [] } ], "base": "urio", "display": "rio", "name": "Realio Network Rio", "symbol": "rio" }, { "description": "Realio Security Token", "denom_units": [ { "denom": "urst", "exponent": 0, "aliases": [ "microrst" ] }, { "denom": "rst", "exponent": 6, "aliases": [] } ], "base": "urst", "display": "rst", "name": "Realio Security Token", "symbol": "rst" }]' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json

# Allocate genesis accounts (cosmos formatted addresses)
realio-networkd add-genesis-account $KEY 10000000000000urio,10000000000000urst --keyring-backend $KEYRING

# Sign genesis transaction
realio-networkd gentx $KEY 1000000000000urio --keyring-backend $KEYRING --chain-id $CHAINID

# Collect genesis tx
realio-networkd collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
realio-networkd validate-genesis

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
realio-networkd start --pruning=nothing $TRACE --log_level $LOGLEVEL --minimum-gas-prices=0.0001urio --json-rpc.api eth,txpool,personal,net,debug,web3