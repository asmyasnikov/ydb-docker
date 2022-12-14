# `ydb-docker` - tool and project for build and run YDB in docker container in single-node configuration

## Environment variables <a name="environ"></a>

`ydb-go-sdk` supports next environment variables  which redefines default behavior of driver

| Name                             | Type      | Default | Description                                                                                                              |
|----------------------------------|-----------|---------|--------------------------------------------------------------------------------------------------------------------------|
| `YDB_SSL_ROOT_CERTIFICATES_FILE` | `string`  |         | path to certificates file                                                                                                |
| `YDB_LOG_SEVERITY_LEVEL`         | `string`  | `quiet` | severity logging level of internal driver logger. Supported: `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `quiet` |
| `YDB_LOG_DETAILS`                | `string`  | `.*`    | regexp for lookup internal logger logs                                                                                   |
| `GRPC_GO_LOG_VERBOSITY_LEVEL`    | `integer` |         | set to `99` to see grpc logs                                                                                             |
| `GRPC_GO_LOG_SEVERITY_LEVEL`     | `string`  |         | set to `info` to see grpc logs                                                                                           |
