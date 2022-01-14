# Use multi-stage build
FROM ubuntu:20.04 as ubuntu-go-base

RUN mkdir /app
WORKDIR /app

RUN apt-get update \
      && apt-get install -y curl \
      && rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION=1.15.12
ENV GOLANG_DOWNLOAD_SHA256=bbdb935699e0b24d90e2451346da76121b2412d30930eabcd80907c230d098b7
ENV GOLANG_ARCHIVE_FILENAME=go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_URL=https://golang.org/dl/$GOLANG_ARCHIVE_FILENAME

RUN curl -sSLO $GOLANG_DOWNLOAD_URL \
      && echo "$GOLANG_DOWNLOAD_SHA256  $GOLANG_ARCHIVE_FILENAME" | sha256sum -c - \
      && tar -C /usr/local -xzf $GOLANG_ARCHIVE_FILENAME \
      && rm $GOLANG_ARCHIVE_FILENAME

ENV PATH=$PATH:/usr/local/go/bin

FROM ubuntu-go-base as kava-rosetta-build

RUN apt-get update \
      && apt-get install -y git make gcc \
      && rm -rf /var/lib/apt/lists/*

ARG kava_node_version=v0.15.0
ENV KAVA_NODE_VERSION=$kava_node_version

RUN git clone https://github.com/kava-labs/kava \
      && cd kava \
      && git checkout $KAVA_NODE_VERSION \
      && make install

COPY . rosetta-kava

RUN cd rosetta-kava \
  && make install

FROM ubuntu:20.04

RUN apt-get update \
      && apt-get install -y supervisor curl \
      && rm -rf /var/lib/apt/lists/*

RUN mkdir /app \
      && mkdir /app/bin
WORKDIR /app

ENV PATH=$PATH:/app/bin

# copy build binaries from build environemtn
COPY --from=kava-rosetta-build /root/go/bin/kvd /app/bin/kvd
COPY --from=kava-rosetta-build /root/go/bin/rosetta-kava /app/bin/rosetta-kava

# copy config templates to automate setup
COPY --from=kava-rosetta-build /app/rosetta-kava/examples /app/templates

# copy scripts to run services
COPY --from=kava-rosetta-build /app/rosetta-kava/conf/start-services.sh /app/bin/start-services.sh
COPY --from=kava-rosetta-build /app/rosetta-kava/conf/kill-supervisord.sh /app/bin/kill-supervisord.sh
COPY --from=kava-rosetta-build /app/rosetta-kava/conf/supervisord.conf /etc/supervisor/conf.d/rosetta-kava.conf

ENV KAVA_RPC_URL=http://localhost:26657

CMD ["/app/bin/start-services.sh"]
