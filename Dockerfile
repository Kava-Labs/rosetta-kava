# Use multi-stage build
FROM golang:1.15 as builder

# Build kava node
RUN git clone https://github.com/kava-labs/kava \
    && cd kava \
    && git checkout v0.14.1 \
    && make install

# Build rosetta-kava service
RUN git clone https://github.com/kava-labs/rosetta-kava \
    && cd rosetta-kava \
    && git fetch origin \
    && git checkout origin/dm-docker-deployment \
    && make install

CMD cd rosetta-kava \
    && ./scripts/start-services.sh
