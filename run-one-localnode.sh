#!/bin/bash
killall realio-networkd || true
rm -rf $HOME/.realio-network
m1="salmon fashion film curve cause palace ancient honey cactus donkey inhale awful resource run junior evil impact border off jacket behave rifle agree eagle"
m2="method kiss layer inherit grain define lecture document corn giraffe galaxy salute ensure mixture release punch duty ridge comfort road cross short review trend"
echo "salmon fashion film curve cause palace ancient honey cactus donkey inhale awful resource run junior evil impact border off jacket behave rifle agree eagle"| realio-networkd keys add val1 --keyring-backend test --recover #--algo ed25519
echo $m2| realio-networkd keys add val2 --keyring-backend test  --recover #--algo ed25519

# realio17eu6f96fk6wqpdajgaqwp70jdxj2jjaxp4regz
val1addr=$(realio-networkd keys show val1  --keyring-backend test -a) 
# # realio1982s57g5kadu2v4ygapne9vj0t7p9d597yr0m8
val2addr=$(realio-networkd keys show val2  --keyring-backend test -a)

# # # init chain
realio-networkd init test --chain-id realionetworktest_3301-1

# Change parameter token denominations to stake
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="stake"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="stake"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="stake"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json
cat $HOME/.realio-network/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="stake"' > $HOME/.realio-network/config/tmp_genesis.json && mv $HOME/.realio-network/config/tmp_genesis.json $HOME/.realio-network/config/genesis.json


echo $val1addr
# Allocate genesis accounts (cosmos formatted addresses)
realio-networkd genesis add-genesis-account realio1jyrr9ga485mzdw6u7w7vcvcmhz8h6zq8w4vxzu 75000000000000000000000000stake --keyring-backend test
realio-networkd genesis add-genesis-account $val2addr 75000000000000000000000000stake --keyring-backend test

# # # # Sign genesis transaction
realio1jyrr9ga485mzdw6u7w7vcvcmhz8h6zq8w4vxzu
# realiovaloper1jyrr9ga485mzdw6u7w7vcvcmhz8h6zq86p0un6
realio-networkd genesis gentx val1  10000000000000000000stake --keyring-backend test --chain-id realionetworktest_3301-1

# # # Collect genesis tx
realio-networkd genesis collect-gentxs

# # # Run this to ensure everything worked and that the genesis file is setup correctly
realio-networkd genesis validate

# # # Start the node (remove the --pruning=nothing flag if historical queries are not needed)
# screen -S realio -t realio -d -m 
realio-networkd start

# sleep 7

# realio-networkd tx bank send $val2 $test2 100000stake  --chain-id realio_3-2 --keyring-backend test --fees 10stake -y #--node tcp://127.0.0.1:26657