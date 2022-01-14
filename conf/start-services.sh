#!/usr/bin/env bash

export START_KAVA_NODE=true

if [ "$MODE" = "offline" ]
then
  START_KAVA_NODE=false
fi

if [ "$START_KAVA_NODE" = "true" ]
then
  mkdir -p /data/kvd/config

  case "$NETWORK" in
    kava-8)
      GENESIS_URL=https://kava-genesis-files.s3.amazonaws.com/kava-8-genesis-migrated-from-block-1878508.json
      ;;
    kava-7)
      GENESIS_URL=https://raw.githubusercontent.com/Kava-Labs/launch/master/kava-7/genesis.json
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

  echo "initializing config..."

  if [ ! -f "/data/kvd/config/app.toml" ]
  then
    cp "/app/templates/$NETWORK/app.toml" /data/kvd/config/app.toml
  fi

  if [ ! -f "/data/kvd/config/config.toml" ]
  then
    cp "/app/templates/$NETWORK/config.toml" /data/kvd/config/config.toml
  fi

  if [ ! -f "/data/kvd/config/genesis.json" ]
  then
    echo "downloading genesis..."
    curl -s "$GENESIS_URL" -o /data/kvd/config/genesis.json
  fi
fi

supervisord -c /etc/supervisor/supervisord.conf
