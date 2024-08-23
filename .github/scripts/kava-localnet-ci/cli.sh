#!/bin/bash

end_idx=($(curl -s --location --request POST 'http://localhost:8000/network/status' \
--header 'Content-Type: application/json' \
--data-raw '{
    "network_identifier": {
        "blockchain": "Kava",
        "network": "kava-localnet"
    }
}' | python3 -c 'import json,sys;obj=json.load(sys.stdin);print(obj["current_block_identifier"]["index"])'))

start_idx=1

echo "start check:data"
echo "start_idx: $start_idx"
echo "end_idx  : $end_idx"
./bin/rosetta-cli --configuration-file rosetta-cli-conf/kava-localnet-ci/config.json check:data --start-block $start_idx --end-block $end_idx
