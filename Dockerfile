# Use multi-stage build
FROM golang:1.15 as builder

ARG KAVANODE_VERSION
ARG NETWORK_ID

RUN ls /data

# Start Kava node
RUN git clone https://github.com/kava-labs/kava \
    && cd kava \
    && git checkout $KAVANODE_VERSION \
    && make install \
    && kvd start --home /data/kvd

# Build Rosetta service
RUN cd .. \
    && git clone https://github.com/kava-labs/rosetta-kava \
    && cd rosetta-kava \
    && git checkout master \
    make install

# Create final container
# FROM ubuntu:latest

# Start Rosetta service once Kava node accepts connections
RUN chmod +x scripts/wait-for-it.sh \
    && scripts/wait-for-it.sh --timeout=0 localhost:26657 \
    && MODE=online NETWORK=$NETWORK_ID PORT=8000 KAVA_RPC_URL=tcp://localhost:26657 rosetta-kava run
