ARG ARCH=amd64

FROM ${ARCH}/debian:stable-slim AS builder

COPY ./artifacts /build

WORKDIR /build

RUN apt update && apt install -y ca-certificates upx 

RUN mkdir -p /build/etc/ssl && cp -r /etc/ssl/certs /build/etc/ssl/certs

RUN chmod +x ./bin/ydbd

# RUN upx ./bin/ydbd

RUN chmod +x ./bin/ydb

# RUN upx ./bin/ydb

RUN chmod +x ./local_ydb

# RUN upx ./local_ydb

COPY ./initialize_local_ydb.sh ./initialize_local_ydb

RUN chmod +x ./initialize_local_ydb

COPY ./health_check.sh ./health_check.sh

RUN chmod +x ./health_check.sh

FROM ${ARCH}/busybox:glibc

COPY --from=builder /build/ /
COPY --from=builder /lib/x86_64-linux-gnu/ /lib/x86_64-linux-gnu/

HEALTHCHECK --interval=1s CMD sh /health_check.sh

CMD ["/initialize_local_ydb"]