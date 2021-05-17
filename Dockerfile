# Use multi-stage build
FROM golang:1.15 as builder

ARG KAVANODE_VERSION
ARG ROSETTA_KAVA_VERSION

# Build kava node
RUN git clone https://github.com/kava-labs/kava \
    && cd kava \
    && git checkout $KAVANODE_VERSION \
    && make install

# Build rosetta-kava service
RUN git clone https://github.com/kava-labs/rosetta-kava \
    && cd rosetta-kava \
    && git fetch origin \
    && git checkout $ROSETTA_KAVA_VERSION \
    && make install

CMD cd rosetta-kava \
    && git fetch origin \
    && git checkout origin/dm-docker-deployment \
    && chmod +x ./scripts/start-services.sh \
    && ./scripts/start-services.sh
