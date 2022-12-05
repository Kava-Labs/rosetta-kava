#!/bin/bash

MODE=online NETWORK=kava-testnet PORT=8000 KAVA_RPC_URL=http://50.16.212.18:26658 nohup go run . run > /dev/null 2>&1 &

sleep 60

block_tip=($(curl -s --location --request POST 'http://localhost:8000/network/status' \
--header 'Content-Type: application/json' \
--data-raw '{
    "network_identifier": {
        "blockchain": "Kava",
        "network": "kava-testnet"
    }
}' | python3 -c 'import json,sys;obj=json.load(sys.stdin);print(obj["current_block_identifier"]["index"])'))

echo "latest block index is", $block_tip

