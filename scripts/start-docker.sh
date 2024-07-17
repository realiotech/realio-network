#!/bin/bash

KEY="dev0"
CHAINID="realionetworklocal_7777-1"
MONIKER="mymoniker"
DATA_DIR=$(mkdir -d -t realionetwork-datadir.XXXXX)
GENESIS=$DATA_DIR/config/genesis.json
TMP_GENESIS=$DATA_DIR/config/tmp_genesis.json

echo "create and add new keys"
./realio-networkd keys add $KEY --home $DATA_DIR --no-backup --chain-id $CHAINID --algo "eth_secp256k1" --keyring-backend test
echo "init RealioNetwork with moniker=$MONIKER and chain-id=$CHAINID"
./realio-networkd init $MONIKER --chain-id $CHAINID --home $DATA_DIR
echo "prepare genesis: Allocate genesis accounts"
./realio-networkd add-genesis-account \
"$(./realio-networkd keys show $KEY -a --home $DATA_DIR --keyring-backend test)" 1000000000000000000ario,1000000000000000000arst \
--home $DATA_DIR --keyring-backend test
echo "prepare genesis: Sign genesis transaction"
./realio-networkd gentx $KEY 1000000000000000000ario--keyring-backend test --home $DATA_DIR --keyring-backend test --chain-id $CHAINID
echo "prepare genesis: Collect genesis tx"
./realio-networkd collect-gentxs --home $DATA_DIR
echo "prepare genesis: Run validate-genesis to ensure everything worked and that the genesis file is setup correctly"
./realio-networkd validate-genesis --home $DATA_DIR

echo "starting RealioNetwork node $i in background ..."
#./realio-networkd start --pruning=nothing --rpc.unsafe \
#--keyring-backend test --home $DATA_DIR \
#>$DATA_DIR/node.log 2>&1 & disown

echo "started RealioNetwork node"
tail -f /dev/null