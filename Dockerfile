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
    && git checkout $ROSETTA_KAVA_VERSION \
    && make install

CMD cd rosetta-kava \
    && git checkout remotes/origin/dm-docker-deployment \
    && chmod +x /rosetta-kava/scripts/start-services.sh \
    && /rosetta-kava/scripts/start-services.sh
