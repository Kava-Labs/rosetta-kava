# Kava Rosetta API

Kava implementation of the Coinbase [Rosetta API](https://www.rosetta-api.org/).

Written in Golang with the [Rosetta Go SDK](https://github.com/coinbase/rosetta-sdk-go).

# Usage

```
make install

MODE=online NETWORK=kava-5.1 PORT=8000 rosetta-kava run
```

# Swagger

Swagger requires a running rosetta-kava service on port 8000.
```
make run-swagger
```
Navigate to [http://localhost:8080](http://localhost:8080).
