#!/bin/bash

./build/realio-networkd tx gov submit-proposal regist_erc20.json --from dev0 --keyring-backend test --chain-id realionetworklocal_7777-1 --gas 600000
sleep 5
./build/realio-networkd query gov proposal 1

./build/realio-networkd tx gov vote 1 yes --from dev0 --keyring-backend test --chain-id realionetworklocal_7777-1

./build/realio-networkd tx gov vote 1 yes --from dev1 --keyring-backend test --chain-id realionetworklocal_7777-1

./build/realio-networkd tx gov vote 1 yes --from dev2 --keyring-backend test --chain-id realionetworklocal_7777-1

./build/realio-networkd query gov proposal 1
