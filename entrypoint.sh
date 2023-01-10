#!/bin/sh

set -ex

if [ ${YDB_USE_IN_MEMORY_PDISKS} ]; then
  YDB_PDISK_PATH="SectorMap:1:64"
else
  YDB_PDISK_PATH="${YDB_DATA_PATH}/ydb.data"
  fallocate -l ${YDB_PDISK_SIZE}G ${YDB_PDISK_PATH}
fi

CPU_CORES=$(grep ^cpu\\scores /proc/cpuinfo | wc -l)

FEATURE_FLAGS=$(echo "${FEATURE_FLAGS}" | sed -e 's/\;/\n  /g; s/^/\n  /g')

mkdir -p ${YDB_DATA_PATH}/
cat << EOF > ${YDB_DATA_PATH}/config.yaml
static_erasure: none
host_configs:
  - drive:
      - path: ${YDB_PDISK_PATH}
        type: SSD
    host_config_id: 1
hosts:
  - host: ${HOSTNAME}
    node_id: 1
    host_config_id: 1
    port: ${IC_PORT}
    walle_location:
      body: 1
      data_center: '1'
      rack: '1'
actor_system_config:
  batch_executor: 2
  executor:
    - name: System
      spin_threshold: 0
      threads: ${CPU_CORES}
      type: BASIC
    - name: User
      spin_threshold: 0
      threads: ${CPU_CORES}
      type: BASIC
    - name: Batch
      spin_threshold: 0
      threads: ${CPU_CORES}
      type: BASIC
    - name: IO
      threads: ${CPU_CORES}
      time_per_mailbox_micro_secs: 100
      type: IO
    - name: IC
      spin_threshold: 10
      threads: ${CPU_CORES}
      time_per_mailbox_micro_secs: 100
      type: BASIC
  io_executor: 3
  scheduler:
    progress_threshold: 10000
    resolution: 256
    spin_threshold: 0
  service_executor:
    - executor_id: 4
      service_name: Interconnect
  sys_executor: 0
  user_executor: 1
blob_storage_config:
  service_set:
    availability_domains: 1
    groups:
      - erasure_species: 0
        group_generation: 0
        group_id: 0
        rings:
          - fail_domains:
              - vdisk_locations:
                  - node_id: 1
                    pdisk_guid: 1
                    pdisk_id: 1
                    vdisk_slot_id: 0
    pdisks:
      - node_id: 1
        path: ${YDB_PDISK_PATH}
        pdisk_category: 1
        pdisk_guid: 1
        pdisk_id: 1
    vdisks:
      - vdisk_id:
          domain: 0
          group_generation: 1
          group_id: 0
          ring: 0
          vdisk: 0
        vdisk_location:
          node_id: 1
          pdisk_guid: 1
          pdisk_id: 1
          vdisk_slot_id: 0
channel_profile_config:
  profile:
    - channel:
        - erasure_species: none
          pdisk_category: 1
          storage_pool_kind: ${STORAGE_POOL_KIND}
        - erasure_species: none
          pdisk_category: 1
          storage_pool_kind: ${STORAGE_POOL_KIND}
        - erasure_species: none
          pdisk_category: 1
          storage_pool_kind: ${STORAGE_POOL_KIND}
      profile_id: 0
    - channel:
        - erasure_species: none
          pdisk_category: 1
          storage_pool_kind: ${STORAGE_POOL_KIND}
        - erasure_species: none
          pdisk_category: 1
          storage_pool_kind: ${STORAGE_POOL_KIND}
        - erasure_species: none
          pdisk_category: 1
          storage_pool_kind: ${STORAGE_POOL_KIND}
        - erasure_species: none
          pdisk_category: 1
          storage_pool_kind: ${STORAGE_POOL_KIND}
        - erasure_species: none
          pdisk_category: 1
          storage_pool_kind: ${STORAGE_POOL_KIND}
        - erasure_species: none
          pdisk_category: 1
          storage_pool_kind: ${STORAGE_POOL_KIND}
        - erasure_species: none
          pdisk_category: 1
          storage_pool_kind: ${STORAGE_POOL_KIND}
      profile_id: 1
domains_config:
  domain:
    - domain_id: 1
      name: local
      scheme_root: 72075186232723360
      storage_pool_types:
        - kind: ${STORAGE_POOL_KIND}
          pool_config:
            box_id: 1
            erasure_species: none
            kind: ${STORAGE_POOL_KIND}
            pdisk_filter:
              - property:
                  - type: SSD
            vdisk_kind: Default
  state_storage:
    - ring:
        node:
          - 1
        nto_select: 1
      ssid: 1
feature_flags:
  enable_mvcc: VALUE_TRUE
  enable_persistent_query_stats: true
  enable_public_api_external_blobs: false
  enable_scheme_transactions_at_scheme_shard: true
  enable_predicate_extract_for_scan_queries: true
  enable_predicate_extract_for_data_queries: true${FEATURE_FLAGS}
grpc_config:
  host: '[::]'
  ca: ${YDB_GRPC_TLS_DATA_PATH}/ca.pem
  cert: ${YDB_GRPC_TLS_DATA_PATH}/cert.pem
  key: ${YDB_GRPC_TLS_DATA_PATH}/key.pem
  services:
    - auth
    - monitoring
    - legacy
    - yql
    - discovery
    - cms
    - locking
    - kesus
    - pq
    - pqcd
    - pqv1
    - datastreams
    - scripting
    - clickhouse_internal
    - rate_limiter
    - analytics
    - export
    - import
    - yq
interconnect_config:
  start_tcp: true
kqpconfig:
  settings:
    - name: _ResultRowsLimit
      value: '1000'
    - name: _KqpYqlSyntaxVersion
      value: '1'
    - name: _KqpAllowNewEngine
      value: 'true'
    - name: _KqpForceNewEngine
      value: 'true'
log_config:
  default_level: ${YDB_DEFAULT_LOG_LEVEL}
  entry: []
  sys_log: false
nameservice_config:
  node:
    - address: ::1
      host: ${HOSTNAME}
      node_id: 1
      port: ${IC_PORT}
      walle_location:
        body: 1
        data_center: '1'
        rack: '1'
net_classifier_config:
  cms_config_timeout_seconds: 30
  net_data_file_path: /ydb_data/netData.tsv
  updater_config:
    net_data_update_interval_seconds: 60
    retry_interval_seconds: 30
pqcluster_discovery_config:
  enabled: false
pqconfig:
  check_acl: false
  cluster_table_path: ''
  clusters_update_timeout_sec: 1
  enabled: true
  meta_cache_timeout_sec: 1
  quoting_config:
    enable_quoting: false
  require_credentials_in_new_protocol: false
  root: ''
  topics_are_first_class_citizen: true
  version_table_path: ''
sqs_config:
  enable_dead_letter_queues: true
  enable_sqs: false
  force_queue_creation_v2: true
  force_queue_deletion_v2: true
  scheme_cache_hard_refresh_time_seconds: 0
  scheme_cache_soft_refresh_time_seconds: 0
system_tablets:
  default_node:
    - 1
  flat_schemeshard:
    - info:
        tablet_id: 72075186232723360
  flat_tx_coordinator:
    - node:
        - 1
  tx_allocator:
    - node:
        - 1
  tx_mediator:
    - node:
        - 1
yandex_query_config:
  audit:
    enabled: false
    uaconfig:
      uri: ''
  checkpoint_coordinator:
    checkpointing_period_millis: 1000
    enabled: true
    max_inflight: 1
    storage:
      endpoint: ''
  common:
    ids_prefix: pt
    use_bearer_for_ydb: true
  control_plane_proxy:
    enabled: true
    request_timeout: 30s
  control_plane_storage:
    available_binding:
      - DATA_STREAMS
      - OBJECT_STORAGE
    available_connection:
      - YDB_DATABASE
      - CLICKHOUSE_CLUSTER
      - DATA_STREAMS
      - OBJECT_STORAGE
      - MONITORING
    enabled: true
    storage:
      endpoint: ''
  db_pool:
    enabled: true
    storage:
      endpoint: ''
  enabled: false
  gateways:
    dq:
      default_settings: []
    enabled: true
    pq:
      cluster_mapping: []
    solomon:
      cluster_mapping: []
  nodes_manager:
    enabled: true
  pending_fetcher:
    enabled: true
  pinger:
    ping_period: 30s
  private_api:
    enabled: true
  private_proxy:
    enabled: true
  resource_manager:
    enabled: true
  token_accessor:
    enabled: true
EOF

cat << EOF > ${YDB_DATA_PATH}/bind_storage_request.txt
ModifyScheme {
  WorkingDir: "/"
  OperationType: ESchemeOpAlterSubDomain
  SubDomain {
    Name: "local"
    StoragePools {
      Name: "${STORAGE_POOL_NAME}"
      Kind: "${STORAGE_POOL_KIND}"
    }
  }
}
EOF

cat << EOF > ${YDB_DATA_PATH}/define_storage_pools_request.txt
Command {
  DefineStoragePool {
    BoxId: 1
    StoragePoolId: 1
    Name: "${STORAGE_POOL_NAME}"
    ErasureSpecies: "none"
    VDiskKind: "Default"
    Kind: "${STORAGE_POOL_KIND}"
    NumGroups: 1
    PDiskFilter {
      Property {
        Type: SSD
      }
      Property {
        Kind: 0
      }
    }
  }
}
EOF

/bin/ydb_certs ${YDB_GRPC_TLS_DATA_PATH}/
/ydbd server --node=1 --ca=${YDB_GRPC_TLS_DATA_PATH}/ca.pem --grpc-port=${GRPC_PORT} --grpcs-port=${GRPC_TLS_PORT} --mon-port=${MON_PORT} --ic-port=${IC_PORT} --yaml-config=${YDB_DATA_PATH}/config.yaml & PID=$!
sleep 1 # TODO: fix it with polling init command
/ydbd -s grpc://localhost:${GRPC_PORT} admin blobstorage config init --yaml-file ${YDB_DATA_PATH}/config.yaml
/ydbd -s grpc://localhost:${GRPC_PORT} admin blobstorage config invoke --proto-file=${YDB_DATA_PATH}/define_storage_pools_request.txt
/ydbd -s grpc://localhost:${GRPC_PORT} db schema execute ${YDB_DATA_PATH}/bind_storage_request.txt
wait $PID
