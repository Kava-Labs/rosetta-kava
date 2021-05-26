# Kava Rosetta API


<p align="center">
  <a href="https://www.rosetta-api.org">
    <img width="90%" alt="Rosetta" src="https://www.rosetta-api.org/img/rosetta_header.png">
  </a>
</p>


Kava implementation of the Coinbase [Rosetta API](https://www.rosetta-api.org/).

Written in Golang with the [Rosetta Go SDK](https://github.com/coinbase/rosetta-sdk-go).

## Features

* Tracking of all native token balance changes for all transaction types
* Stateless, offline transaction construction
* Historical balance lookup and reconciliation

## Prerequisites

To run `rosetta-kava`, [docker](https://docs.docker.com/get-docker/) is required.


## System Requirements

`rosetta-kava` has been tested on an [AWS c5.2xlarge instance](https://aws.amazon.com/ec2/instance-types/c5). We recommend 8 vCPU, 16GB of RAM, and at least 1TB of storage for running a dockerized `rosetta-kava` node.

## Usage

As specified in the Rosetta API, the `rosetta-kava` implementation is deployable via Docker and supports running via either an `online` or `offline` mode.

## Install

### Mainnet

The following commands will build a docker container named `rosetta-kava` and configure the container for running on the `kava-7` mainnet.

```
mkdir -p kava-data/kvd/config

cp examples/kava-7/app.toml kava-data/kvd/config/app.toml
cp examples/kava-7/config.toml kava-data/kvd/config/config.toml
curl https://raw.githubusercontent.com/Kava-Labs/launch/master/kava-7/genesis.json > kava-data/kvd/config/genesis.json

docker build . -t rosetta-kava
docker run -it -e "MODE=online" -e "NETWORK=kava-7" -e "PORT=8000" -v "$PWD/kava-data:/data" -p 8000:8000 -p 26656:26656 rosetta-kava
```

To run in offline mode:

```
docker run -it -e "MODE=offline" -e "NETWORK=kava-7" -e "PORT=8000" -p 8000:8000 rosetta-kava
```


### Testnet

The following commands will build a docker container named `rosetta-kava` and configure the container for running on the `kava-testnet-12000` testnet.

```
mkdir -p kava-data/kvd/config

cp examples/kava-testnet-12000/app.toml kava-data/kvd/config/app.toml
cp examples/kava-testnet-12000/config.toml kava-data/kvd/config/config.toml
curl https://raw.githubusercontent.com/Kava-Labs/kava-testnets/master/12000/genesis.json > kava-data/kvd/config/genesis.json

docker build . -t rosetta-kava --build-arg kava_node_version=v0.14.0-rc1
docker run -it -e "MODE=online" -e "NETWORK=kava-testnet-12000" -e "PORT=8000" -v "$PWD/kava-data:/data" -p 8000:8000 -p 26656:26656 rosetta-kava
```

To run in offline mode:

```
docker run -it -e "MODE=offline" -e "NETWORK=kava-testnet-12000" -e "PORT=8000" -p 8000:8000 rosetta-kava
```

# Swagger

Swagger requires a running rosetta-kava service on port 8000.
```
make run-swagger
```
Navigate to [http://localhost:8080](http://localhost:8080).
