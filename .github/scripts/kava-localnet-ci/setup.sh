#!/bin/bash

make run-local &
make run-local-offline &

sleep 5

block_tip=($(curl -s --location --request POST 'http://localhost:8000/network/status' \
--header 'Content-Type: application/json' \
--data-raw '{
    "network_identifier": {
        "blockchain": "Kava",
        "network": "kava-localnet"
    }
}' | python3 -c 'import json,sys;obj=json.load(sys.stdin);print(obj["current_block_identifier"]["index"])'))

echo "latest block index is", $block_tip

