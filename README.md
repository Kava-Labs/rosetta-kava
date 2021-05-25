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

### Building docker container
TODO


## System Requirements

`rosetta-kava` has been tested on an [AWS c5.2xlarge instance](https://aws.amazon.com/ec2/instance-types/c5). We recommend 8 vCPU, 16GB of RAM, and at least 1TB of storage for running a dockerized `rosetta-kava` node.

## Usage


```
make install

MODE=online NETWORK=kava-7 PORT=8000 rosetta-kava run
```

# Swagger

Swagger requires a running rosetta-kava service on port 8000.
```
make run-swagger
```
Navigate to [http://localhost:8080](http://localhost:8080).
