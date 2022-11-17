#!/bin/sh

set -e
set -x

/bin/ydb -e grpcs://localhost:${GRPC_TLS_PORT:-2135} --ca-file /ydb_certs/ca.pem -d /local scheme ls /local
/bin/ydb -e grpcs://localhost:${GRPC_TLS_PORT:-2135} --ca-file /ydb_certs/ca.pem -d /local table query execute -q 'create table `/local/.sys_health/test` (key int32, value utf8, primary key(key));' -t scheme