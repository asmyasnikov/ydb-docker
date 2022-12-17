ARG ARCH=amd64

FROM ${ARCH}/debian:stable AS builder

ARG COMPRESS_BINARIES=false

RUN apt update && apt install -y ca-certificates upx

RUN mkdir -p /build/etc/ssl && cp -r /etc/ssl/certs /build/etc/ssl/certs

COPY ./artifacts /build

WORKDIR /build

RUN chmod +x ./bin/ydbd

RUN $COMPRESS_BINARIES && upx ./bin/ydbd || :

RUN mv ./bin/ydbd ./

RUN chmod +x ./bin/ydb

RUN $COMPRESS_BINARIES && upx ./bin/ydb || :

RUN mv ./bin/ydb ./

RUN chmod +x ./entrypoint

RUN $COMPRESS_BINARIES && upx ./entrypoint || :

RUN mkdir -p ./root/ydb/bin/

RUN echo '{"check_version":false}' > ./root/ydb/bin/config.json

COPY ./health_check.sh ./health_check.sh

RUN chmod +x ./health_check.sh

FROM ${ARCH}/busybox:glibc

COPY --from=builder /build/ /
COPY --from=builder /lib/x86_64-linux-gnu/ /lib/x86_64-linux-gnu/

HEALTHCHECK --interval=1s CMD sh /health_check.sh

CMD ["/entrypoint"]