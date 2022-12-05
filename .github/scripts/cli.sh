#!/bin/bash

# downloading cli
curl -sSfL https://raw.githubusercontent.com/coinbase/rosetta-cli/master/scripts/install.sh | sh -s

end_idx=($(curl -s --location --request POST 'http://localhost:8000/network/status' \
--header 'Content-Type: application/json' \
--data-raw '{
    "network_identifier": {
        "blockchain": "Kava",
        "network": "kava-testnet"
    }
}' | python3 -c 'import json,sys;obj=json.load(sys.stdin);print(obj["current_block_identifier"]["index"])'))

lastest_X_blocks=10
start_idx=$(($end_idx - $lastest_X_blocks))

echo "start check:data"
./bin/rosetta-cli --configuration-file rosetta-cli-conf/kava-testnet-ci/config.json check:data --start-block $start_idx --end-block $end_idx
