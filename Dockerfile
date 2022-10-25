ARG ARCH=amd64

FROM ${ARCH}/debian:stable-slim AS builder

WORKDIR /build

# COPY ./configs/config.yml ydb_data/kikimr_configs/config.yaml
COPY ./ydb_certs/ ./ydb_certs
COPY ./entrypoint.sh ./entrypoint.sh

RUN apt update && apt install -y ca-certificates wget

ARG VERSION=22.4.31

RUN wget https://binaries.ydb.tech/release/22.4.31/ydbd-${VERSION}-linux-amd64.tar.gz && tar --strip-components=1 -xvzf ydbd-${VERSION}-linux-amd64.tar.gz && rm -f ydbd-${VERSION}-linux-amd64.tar.gz

RUN chmod +x ./entrypoint.sh

# RUN cp /lib/x86_64-linux-gnu/libdl.so.2 ./lib/
# RUN cp /lib/x86_64-linux-gnu/librt.so.2 ./lib/

FROM ${ARCH}/debian:stable-slim
# FROM ${ARCH}/busybox:glibc

COPY --from=builder /build/ /

ENTRYPOINT [ "/entrypoint.sh" ]

CMD sh -c "/bin/ydbd server --node=1 --ca=/ydb_certs/ca.pem --grpcs-port=2135 --yaml-config=/ydb_data/kikimr_configs/config.yaml --grpc-port=2136 --mon-port=8765 --ic-port=19001"