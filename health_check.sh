#!/bin/sh

set -e
set -x

/ydb -e grpc://localhost:${GRPC_TLS_PORT:-2136} -d /local scheme ls /local
/ydb -e grpc://localhost:${GRPC_TLS_PORT:-2136} -d /local table query execute -q 'create table `/local/.sys_health/test` (key int32, value utf8, primary key(key));' -t scheme