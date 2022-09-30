#!/usr/bin/env bash

export START_KAVA_NODE=true

if [ "$MODE" = "offline" ]
then
  START_KAVA_NODE=false
fi

if [ "$START_KAVA_NODE" = "true" ]
then
  mkdir -p /data/kava/config

  case "$NETWORK" in
    kava-mainnet)
      GENESIS_URL=https://kava-genesis-files.s3.amazonaws.com/kava_2222-10/genesis.json
      ;;
    kava-testnet)
      GENESIS_URL=https://raw.githubusercontent.com/Kava-Labs/kava-testnets/master/16000/genesis.json
      ;;
    kava-9)
      GENESIS_URL=https://kava-genesis-files.s3.amazonaws.com/kava-9/genesis.json
      ;;
    kava-8)
      GENESIS_URL=https://kava-genesis-files.s3.amazonaws.com/kava-8-genesis-migrated-from-block-1878508.json
      ;;
    kava-7)
      GENESIS_URL=https://raw.githubusercontent.com/Kava-Labs/launch/master/kava-7/genesis.json
      ;;
    kava-testnet-14000)
      GENESIS_URL=https://raw.githubusercontent.com/Kava-Labs/kava-testnets/master/14000/genesis.json
      ;;
    kava-testnet-13000)
      GENESIS_URL=https://raw.githubusercontent.com/Kava-Labs/kava-testnets/master/13000/genesis.json
      ;;
    kava-testnet-12000)
      GENESIS_URL=https://raw.githubusercontent.com/Kava-Labs/kava-testnets/master/12000/genesis.json
      ;;
    *)
      echo "unknown network"
      exit 1
      ;;
  esac

  echo "initializing cosmovisor..."

  if [ ! -d "/data/kava/cosmovisor" ]
  then
    cp -r /app/cosmovisor /data/kava/cosmovisor
  else
    cp -Tr /app/cosmovisor/upgrades  /data/kava/cosmovisor/upgrades
  fi

  if [ ! -d "/data/kava/data" ]
  then
    mkdir /data/kava/data
  fi

  echo "initializing config..."

  if [ ! -f "/data/kava/config/app.toml" ]
  then
    cp "/app/templates/$NETWORK/app.toml" /data/kava/config/app.toml
  fi

  if [ ! -f "/data/kava/config/config.toml" ]
  then
    cp "/app/templates/$NETWORK/config.toml" /data/kava/config/config.toml
  fi

  if [ ! -f "/data/kava/config/genesis.json" ]
  then
    echo "downloading genesis..."
    curl -s "$GENESIS_URL" -o /data/kava/config/genesis.json
  fi
fi

supervisord -c /etc/supervisor/supervisord.conf
