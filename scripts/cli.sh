#!/bin/bash

END_INDEX=$(($START_INDEX+$COUNT))
echo "index ranges", $START_INDEX, $END_INDEX

MODE=online NETWORK=kava-testnet PORT=8000 KAVA_RPC_URL=http://50.16.212.18:26658 nohup go run . run > /dev/null 2>&1 &

# downloading cli
curl -sSfL https://raw.githubusercontent.com/coinbase/rosetta-cli/master/scripts/install.sh | sh -s

sleep 180

echo "start check:data"
./bin/rosetta-cli --configuration-file rosetta-cli-conf/kava-testnet/config.json check:data --start-block $START_INDEX --end-block $END_INDEX
