#!/bin/sh

mkdir -p /ydb_data

if [ -z "$YDB_GRPC_ENABLE_TLS" ]; then
  YDB_GRPC_ENABLE_TLS="true"
fi

if [ -z "$GRPC_TLS_PORT" ]; then
  GRPC_TLS_PORT=${GRPC_TLS_PORT:-2135}
fi

if [ -z "$GRPC_PORT" ]; then
  GRPC_PORT=${GRPC_PORT:-2136}
fi

if [ -z "$YDB_GRPC_TLS_DATA_PATH" ]; then
  YDB_GRPC_TLS_DATA_PATH="/ydb_certs"
fi

if [ -z "$YDB_PDISK_CATEGORY_TYPE" ]; then
  YDB_PDISK_CATEGORY_TYPE="SSD"
fi

if [ -z "$YDB_USE_IN_MEMORY_PDISKS" ]; then
  YDB_PDISK_PATH="/ydb_data/ydb.data"
  if [ -z "$YDB_PDISK_SIZE" ]; then
    YDB_PDISK_SIZE="80G"
  fi
  fallocate -l ${YDB_PDISK_SIZE} ${YDB_PDISK_PATH}
  YDB_PDISK_CATEGORY=0
else
  YDB_PDISK_PATH="SectorMap:1:64"
  YDB_PDISK_CATEGORY=1
fi

if [ -z "$YDB_INTERCONNECT_PORT" ]; then
  YDB_INTERCONNECT_PORT="19001"
fi

YDB_PDISK_CATEGORY_TYPE_KIND=$(echo $YDB_PDISK_CATEGORY_TYPE | tr '[:upper:]' '[:lower:]')

cat << EOF > /ydb_data/config.yaml
static_erasure: none
host_configs:
- drive:
  - path: ${YDB_PDISK_PATH}
    type: ${YDB_PDISK_CATEGORY_TYPE}
  host_config_id: 1
hosts:
- host: localhost
  host_config_id: 1
  port: ${YDB_INTERCONNECT_PORT}
  walle_location:
    body: 1
    data_center: '1'
    rack: '1'
domains_config:
  domain:
  - name: local
    storage_pool_types:
    - kind: ${YDB_PDISK_CATEGORY_TYPE_KIND}
      pool_config:
        box_id: 1
        erasure_species: none
        kind: ${YDB_PDISK_CATEGORY_TYPE_KIND}
        pdisk_filter:
        - property:
          - type: ${YDB_PDISK_CATEGORY_TYPE}
        vdisk_kind: Default
  state_storage:
  - ring:
      node:
      - 1
      nto_select: 1
    ssid: 1
table_service_config:
  sql_version: 1
actor_system_config:
  executor:
  - name: System
    spin_threshold: 0
    threads: 2
    type: BASIC
  - name: User
    spin_threshold: 0
    threads: 3
    type: BASIC
  - name: Batch
    spin_threshold: 0
    threads: 2
    type: BASIC
  - name: IO
    threads: 1
    time_per_mailbox_micro_secs: 100
    type: IO
  - name: IC
    spin_threshold: 10
    threads: 1
    time_per_mailbox_micro_secs: 100
    type: BASIC
  scheduler:
    progress_threshold: 10000
    resolution: 256
    spin_threshold: 0
blob_storage_config:
  service_set:
    groups:
    - erasure_species: none
      rings:
      - fail_domains:
        - vdisk_locations:
          - node_id: 1
            path: ${YDB_PDISK_PATH}
            pdisk_category: ${YDB_PDISK_CATEGORY_TYPE}
channel_profile_config:
  profile:
  - channel:
    - erasure_species: none
      pdisk_category: ${YDB_PDISK_CATEGORY}
      storage_pool_kind: ${YDB_PDISK_CATEGORY_TYPE_KIND}
    - erasure_species: none
      pdisk_category: ${YDB_PDISK_CATEGORY}
      storage_pool_kind: ${YDB_PDISK_CATEGORY_TYPE_KIND}
    - erasure_species: none
      pdisk_category: ${YDB_PDISK_CATEGORY}
      storage_pool_kind: ${YDB_PDISK_CATEGORY_TYPE_KIND}
    profile_id: 0
grpc_config:
  host: '[::]'
  ca: ${YDB_GRPC_TLS_DATA_PATH}/ca.pem
  cert: ${YDB_GRPC_TLS_DATA_PATH}/cert.pem
  key: ${YDB_GRPC_TLS_DATA_PATH}/key.pem
EOF

# disable checks of updates cli
/bin/ydb version --disable-checks

# start storage process
/bin/ydbd server --node=1 --ca=${YDB_GRPC_TLS_DATA_PATH}/ca.pem --grpc-port=${GRPC_PORT} --grpcs-port=${GRPC_TLS_PORT} --yaml-config=/ydb_data/config.yaml --mon-port=8765 --ic-port=${YDB_INTERCONNECT_PORT} &

# wait for start
sleep 3

# initialize storage
/bin/ydbd -s grpc://localhost:2136 admin blobstorage config init --yaml-file /ydb_data/config.yaml

# register database
/bin/ydbd -s grpc://localhost:2136 admin database /local create ${YDB_PDISK_CATEGORY_TYPE_KIND}:1

# start dynnode process
/bin/ydbd server --yaml-config /ydb_data/config.yaml  --tenant /local --node-broker localhost:2136 --grpc-port 31001 --ic-port 31002 --mon-port 31003
