ARG ARCH=amd64

FROM ${ARCH}/debian:stable-slim AS builder

WORKDIR /build

RUN apt update && apt install -y ca-certificates wget upx

ARG VERSION=22.4.31

RUN wget https://binaries.ydb.tech/release/${VERSION}/ydbd-${VERSION}-linux-amd64.tar.gz && tar --strip-components=1 -xvzf ydbd-${VERSION}-linux-amd64.tar.gz && rm -f ydbd-${VERSION}-linux-amd64.tar.gz && upx ./bin/ydbd

ARG CLI_VERSION=2.0.0

RUN wget https://storage.yandexcloud.net/yandexcloud-ydb/release/${CLI_VERSION}/linux/amd64/ydb && chmod +x ydb && mv ydb ./bin/ && upx ./bin/ydb

FROM ${ARCH}/busybox:glibc

COPY ./ydb_certs/ ./ydb_certs

COPY ./entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

COPY ./healthcheck.sh ./healthcheck.sh
RUN chmod +x ./healthcheck.sh

COPY --from=builder /build/ /
COPY --from=builder /lib/x86_64-linux-gnu/ /lib/x86_64-linux-gnu/

HEALTHCHECK --interval=1s CMD sh /healthcheck.sh

CMD [ "/entrypoint.sh" ]