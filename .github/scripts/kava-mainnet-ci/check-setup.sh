#!/bin/bash

block_tip=($(curl -s --location --request POST 'http://localhost:8080/network/status' \
--header 'Content-Type: application/json' \
--data-raw '{
    "network_identifier": {
        "blockchain": "Kava",
        "network": "kava-mainnet"
    }
}' | python3 -c 'import json,sys;obj=json.load(sys.stdin);print(obj["current_block_identifier"]["index"])'))

echo "latest block index is" $block_tip
