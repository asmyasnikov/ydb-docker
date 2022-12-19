# `ydb-docker` - tool and project for build and run YDB in docker container in single-node configuration

## Build docker image

1. Non-compressed binaries
    ```shell
    docker build -t amyasnikov/ydb:latest .
    ```
2. Compressed binaries
    ```shell
    docker build -t amyasnikov/ydb:slim --build-arg COMPRESS_BINARIES=true .
    ```

## Environment variables for run docker container

| Name                       | Type      | Default       | Description                  |
|----------------------------|-----------|---------------|------------------------------|
| `YDB_USE_IN_MEMORY_PDISKS` | `boolean` | `false`       | run ydb with in-memory pdisk |
| `YDB_GRPC_TLS_DATA_PATH`   | `string`  | `/ydb_certs/` | certificates directory path  |
| `YDB_DATA_PATH`            | `string`  | `/ydb_data/`  | working directory            |
| `YDB_DEFAULT_LOG_LEVEL`    | `string`  | `5`           | log level of ydb             |
| `GRPC_PORT`                | `integer` | `2136`        | grpc port                    |
| `GRPC_TLS_PORT`            | `integer` | `2135`        | secure grpc port             |
| `MON_PORT`                 | `integer` | `8765`        | port of embedded UI          |
| `IC_PORT`                  | `integer` | `19001`       | port of interconnect         |
| `YDB_PDISK_SIZE`           | `integer` | `80`          | pdisk size in `GB`           |
| `STORAGE_POOL_KIND`        | `string`  | `ssd`         | storage pool kind            |
| `STORAGE_POOL_NAME`        | `string`  | `local`       | storage pool name            |
