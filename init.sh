KEY="eduardo"
KEY2="realio-val-1"
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
make clean
make install

# Set moniker and chain-id for Realio Network (Moniker can be anything, chain-id must be an integer)
realio-networkd init $MONIKER --chain-id $CHAINID

# Set client config
realio-networkd config keyring-backend $KEYRING --home $HOMEDIR
realio-networkd config chain-id $CHAINID --home $HOMEDIR

# if $KEY exists it should be deleted
realio-networkd keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --home $HOMEDIR
realio-networkd keys add $KEY2 --keyring-backend $KEYRING --algo $KEYALGO --home $HOMEDIR

# Change parameter token denominations to ario
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="ario,arst"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="ario"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="ario"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="ario"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="ario"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["inflation"]["params"]["mint_denom"]="ario"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json

# Add denom metadata for rio and rst
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["bank"]["denom_metadata"]=[{ "description": "The native token of the Realio Network", "denom_units": [ { "denom": "ario", "exponent": 0, "aliases": [ "attorio" ] }, { "denom": "rio", "exponent": 18, "aliases": [] } ], "base": "ario", "display": "rio", "name": "Realio Network Rio", "symbol": "rio" }, { "description": "Realio Security Token", "denom_units": [ { "denom": "arst", "exponent": 0, "aliases": [ "attorst" ] }, { "denom": "rst", "exponent": 18, "aliases": [] } ], "base": "arst", "display": "rst", "name": "Realio Security Token", "symbol": "rst" }]' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json

# Allocate genesis accounts (cosmos formatted addresses)
realio-networkd add-genesis-account $KEY 10000000000000000000000000ario,5000000000000000000000000arst --keyring-backend $KEYRING
realio-networkd add-genesis-account $KEY2 1100000000000000000000000ario,1000000000000000000000000arst --keyring-backend $KEYRING

# Sign genesis transaction
realio-networkd gentx $KEY2 1000000000000000000000000ario --keyring-backend $KEYRING --chain-id $CHAINID

# Collect genesis tx
realio-networkd collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
realio-networkd validate-genesis

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
realio-networkd start --pruning=nothing $TRACE --log_level $LOGLEVEL --minimum-gas-prices=0.0001ario --json-rpc.api eth,txpool,personal,net,debug,web3