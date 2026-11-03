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
docker build . -f Dockerfile.mainnet -t rosetta-kava-mainnet
docker run -it -e "MODE=online" -e "NETWORK=kava-mainnet" -e "PORT=8000" -v "$PWD/kava-data:/data" -p 8000:8000 -p 26656:26656 rosetta-kava-mainnet
```

To run in offline mode:

```
docker run -it -e "MODE=offline" -e "NETWORK=kava-mainnet" -e "PORT=8000" -p 8000:8000 rosetta-kava-mainnet
```

### Testnet

The following commands will build a docker container named `rosetta-kava` and configure the container for running on the `kava-testnet` network.

```
docker build . -f Dockerfile.testnet -t rosetta-kava-testnet
docker run -it -e "MODE=online" -e "NETWORK=kava-testnet" -e "PORT=8000" -v "$PWD/kava-data:/data" -p 8000:8000 -p 26656:26656 rosetta-kava-testnet
```

To run in offline mode:

```
docker run -it -e "MODE=offline" -e "NETWORK=kava-testnet" -e "PORT=8000" -p 8000:8000 rosetta-kava-testnet
```

### Client ID header

`rosetta-kava` can send a client ID header on its calls to the Kava RPC node, for operators that run it behind a node gateway that attributes or requires a client ID. The feature is off unless `CLIENT_ID_HEADER` is set.

| Variable | Description |
| --- | --- |
| `CLIENT_ID_HEADER` | Header name, for example `x-client-id`. Node calls made while serving a request carry that request's value of the header. A request without the header sends none. |
| `CLIENT_ID_DEFAULT` | Optional. Value sent on node calls made outside an incoming request. Requires `CLIENT_ID_HEADER`. |

```
docker run -it -e "MODE=online" -e "NETWORK=kava-mainnet" -e "PORT=8000" -e "KAVA_RPC_URL=https://rpc.example.com:443" -e "CLIENT_ID_HEADER=x-client-id" -p 8000:8000 rosetta-kava-mainnet
```

Only that one header is forwarded. Header names that would break the node call (`host`, `content-length`, `content-type`, `transfer-encoding`, `connection`, `upgrade`, `accept`, `sec-websocket-*`) are refused at startup. Header values are never logged.

The header is for attribution, not authorization: callers should be authenticated before they reach `rosetta-kava`.

# Swagger

Swagger requires a running rosetta-kava service on port 8000.
```
make run-swagger
```
Navigate to [http://localhost:8080](http://localhost:8080).
