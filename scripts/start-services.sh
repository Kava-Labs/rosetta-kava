#!/bin/bash

set_access_permissions() {
    chmod +x scripts/wait-for-it.sh
}

start_kava_node()
{
    echo "Starting Kava node"

    kvd unsafe-reset-all
    kvd start --home /data/kvd

    scripts/wait-for-it.sh --timeout=60 localhost:26657
    sleep 5
}

start_rosetta_service() {
    echo "Starting Rosetta service"
    MODE=online NETWORK=$NETWORK_ID PORT=8000 KAVA_RPC_URL=tcp://localhost:26657 rosetta-kava run
}

set_access_permissions
start_kava_node
start_rosetta_service

sleep 10
echo "System has been initialized"
wait
