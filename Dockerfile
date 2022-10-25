ARG ARCH=amd64

FROM ${ARCH}/debian:stable-slim AS builder

WORKDIR /build

# COPY ./configs/config.yml ydb_data/kikimr_configs/config.yaml
COPY ./ydb_certs/ ./ydb_certs
COPY ./entrypoint.sh ./entrypoint.sh

RUN apt update && apt install -y ca-certificates wget

ARG VERSION=22.4.31

RUN wget https://binaries.ydb.tech/release/${VERSION}/ydbd-${VERSION}-linux-amd64.tar.gz && tar --strip-components=1 -xvzf ydbd-${VERSION}-linux-amd64.tar.gz && rm -f ydbd-${VERSION}-linux-amd64.tar.gz

RUN chmod +x ./entrypoint.sh

FROM ${ARCH}/busybox:glibc

COPY --from=builder /build/ /
COPY --from=builder /lib/x86_64-linux-gnu/ /lib/x86_64-linux-gnu/

ENTRYPOINT [ "/entrypoint.sh" ]