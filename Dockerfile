ARG ARCH=amd64

FROM ${ARCH}/debian:stable AS builder

ARG COMPRESS_BINARIES=false

WORKDIR /build

# copy ydb binaries from host to image
COPY ./artifacts ./

# install nessesarry tools and packages
RUN apt update && apt install -y upx

# ydbd and ydb (cli) nessesarry libraries
RUN mkdir -p ./lib && \
    cp /lib/x86_64-linux-gnu/libdl.so* ./lib/ && \
    cp /lib/x86_64-linux-gnu/librt.so* ./lib/

# make executable ydbd
RUN chmod +x ./bin/ydbd

# compress ydbd with upx
RUN $COMPRESS_BINARIES && upx ./bin/ydbd || true

# move ydbd to root
RUN ln -sf /bin/ydbd ./ydbd

# make executable ydb cli
RUN chmod +x ./bin/ydb

# compress ydb cli with upx
RUN $COMPRESS_BINARIES && upx ./bin/ydb || true

# move ydbd to root
RUN ln -sf /bin/ydb ./ydb

# disable check updates of ydb cli
RUN mkdir -p ./root/ydb/bin/ && echo '{"check_version":false}' > ./root/ydb/bin/config.json

# copy entrypoint.sh from host to build dir
COPY ./entrypoint.sh ./entrypoint.sh

# make executable entrypoint.sh
RUN chmod +x ./entrypoint.sh

# copy health_check.sh from host to build dir
COPY ./health_check.sh ./health_check.sh

# make executable health_check.sh
RUN chmod +x ./health_check.sh

# make executable ydb_certs
RUN chmod +x ./bin/ydb_certs

# compress ydb_certs with upx
RUN $COMPRESS_BINARIES && upx ./bin/ydb_certs || true

# busybox is lightweight image with POSIX tools
FROM ${ARCH}/busybox

# copy data from build image
COPY --from=builder /build/ /

# define environment variables with default values
ENV YDB_USE_IN_MEMORY_PDISKS=false
ENV YDB_GRPC_TLS_DATA_PATH=/ydb_certs
ENV YDB_DATA_PATH=/ydb_data
ENV YDB_DEFAULT_LOG_LEVEL=5
ENV GRPC_PORT=2136
ENV GRPC_TLS_PORT=2135
ENV MON_PORT=8765
ENV IC_PORT=19001
ENV YDB_PDISK_SIZE=80
ENV STORAGE_POOL_KIND=ssd
ENV STORAGE_POOL_NAME=local

HEALTHCHECK --interval=1s CMD sh /health_check.sh

CMD ["/entrypoint.sh"]