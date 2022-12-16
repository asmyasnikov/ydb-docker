ARG ARCH=amd64

ARG BINARIES_TYPE=slim

FROM ${ARCH}/debian:stable AS builder

RUN apt update && apt install -y ca-certificates upx

RUN mkdir -p /build/etc/ssl && cp -r /etc/ssl/certs /build/etc/ssl/certs

COPY ./artifacts /build

WORKDIR /build

RUN chmod +x ./bin/ydbd

RUN if [[ "$BINARIES_TYPE" = "slim" ]] ; \
    then \
      upx ./bin/ydbd ; \
    fi

RUN mv ./bin/ydbd ./

RUN chmod +x ./bin/ydb

RUN if [[ "$BINARIES_TYPE" = "slim" ]] ; \
    then \
      upx ./bin/ydb ; \
    fi

RUN mv ./bin/ydb ./

RUN chmod +x ./entrypoint

RUN if [[ "$BINARIES_TYPE" = "slim" ]] ; \
    then \
    upx ./entrypoint ; \
    fi

RUN mkdir -p ./root/ydb/bin/

RUN echo '{"check_version":false}' > ./root/ydb/bin/config.json

COPY ./health_check.sh ./health_check.sh

RUN chmod +x ./health_check.sh

FROM ${ARCH}/busybox:glibc

COPY --from=builder /build/ /
COPY --from=builder /lib/x86_64-linux-gnu/ /lib/x86_64-linux-gnu/

HEALTHCHECK --interval=1s CMD sh /health_check.sh

CMD ["/entrypoint"]