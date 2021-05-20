#!/bin/bash
set -eu

start_kava_node()
{
    echo "Starting Kava node"

    kvd start --home /data/kvd &
    scripts/wait-for-it.sh --timeout=600 localhost:26657 # 10 minute timeout
    sleep 5
}

start_rosetta_service() {
    echo "Starting Rosetta service"
    MODE=online NETWORK=testing PORT=8000 KAVA_RPC_URL=tcp://localhost:26657 rosetta-kava run&
    scripts/wait-for-it.sh --timeout=60 localhost:8000
}

start_kava_node
start_rosetta_service

sleep 10
echo "System has been initialized"
wait
