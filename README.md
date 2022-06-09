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

`rosetta-kava` has been tested on an [AWS c5.2xlarge instance](https://aws.amazon.com/ec2/instance-types/c5). We recommend 8 vCPU, 16GB of RAM, and at least 2TB of storage for running a dockerized `rosetta-kava` node.

## Usage

As specified in the Rosetta API, the `rosetta-kava` implementation is deployable via Docker and supports running via either an `online` or `offline` mode.

## Install

### Mainnet

The following commands will build a docker container named `rosetta-kava` and configure the container for running on the `kava-mainnet` network.

```
docker build . -t rosetta-kava
docker run -it -e "MODE=online" -e "NETWORK=kava-mainnet" -e "PORT=8000" -v "$PWD/kava-data:/data" -p 8000:8000 -p 26656:26656 rosetta-kava
```

To run in offline mode:

```
docker run -it -e "MODE=offline" -e "NETWORK=kava-mainnet" -e "PORT=8000" -p 8000:8000 rosetta-kava
```

### Testnet

The following commands will build a docker container named `rosetta-kava` and configure the container for running on the `kava-testnet` network.

```
docker build . -t rosetta-kava
docker run -it -e "MODE=online" -e "NETWORK=kava-testnet" -e "PORT=8000" -v "$PWD/kava-data:/data" -p 8000:8000 -p 26656:26656 rosetta-kava
```

To run in offline mode:

```
docker run -it -e "MODE=offline" -e "NETWORK=kava-testnet" -e "PORT=8000" -p 8000:8000 rosetta-kava
```

# Swagger

Swagger requires a running rosetta-kava service on port 8000.
```
make run-swagger
```
Navigate to [http://localhost:8080](http://localhost:8080).
