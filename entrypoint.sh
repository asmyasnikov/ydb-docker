#!/bin/sh

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

if [ -z "$YDB_USE_IN_MEMORY_PDISKS" ]; then
  YDB_PDISK_PATH="/ydb_data/pdisk1.data"
else
  YDB_PDISK_PATH="SectorMap:1:64"
fi

mkdir -p /ydb_data

cat << EOF > /ydb_data/config.yaml
actor_system_config:
  batch_executor: 2
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
      pdisk_category: 0
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
      pdisk_category: 0
      storage_pool_kind: hdd
    - erasure_species: none
      pdisk_category: 0
      storage_pool_kind: hdd
    - erasure_species: none
      pdisk_category: 0
      storage_pool_kind: hdd
    profile_id: 0
  - channel:
    - erasure_species: none
      pdisk_category: 0
      storage_pool_kind: hdd
    - erasure_species: none
      pdisk_category: 0
      storage_pool_kind: hdd
    - erasure_species: none
      pdisk_category: 0
      storage_pool_kind: hdd
    - erasure_species: none
      pdisk_category: 0
      storage_pool_kind: hdd
    - erasure_species: none
      pdisk_category: 0
      storage_pool_kind: hdd
    - erasure_species: none
      pdisk_category: 0
      storage_pool_kind: hdd
    - erasure_species: none
      pdisk_category: 0
      storage_pool_kind: hdd
    profile_id: 1
domains_config:
  domain:
  - domain_id: 1
    name: local
    scheme_root: 72075186232723360
    storage_pool_types:
    - kind: hdd
      pool_config:
        box_id: 1
        erasure_species: none
        kind: hdd
        pdisk_filter:
        - property:
          - type: ROT
        vdisk_kind: Default
    - kind: hdd1
      pool_config:
        box_id: 1
        erasure_species: none
        kind: hdd
        pdisk_filter:
        - property:
          - type: ROT
        vdisk_kind: Default
    - kind: hdd2
      pool_config:
        box_id: 1
        erasure_species: none
        kind: hdd
        pdisk_filter:
        - property:
          - type: ROT
        vdisk_kind: Default
    - kind: hdde
      pool_config:
        box_id: 1
        encryption_mode: 1
        erasure_species: none
        kind: hdd
        pdisk_filter:
        - property:
          - type: ROT
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
grpc_config:
  ca: ${YDB_GRPC_TLS_DATA_PATH}/ca.pem
  cert: ${YDB_GRPC_TLS_DATA_PATH}/cert.pem
  host: '[::]'
  key: ${YDB_GRPC_TLS_DATA_PATH}/key.pem
  services:
  - experimental
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
  - s3_internal
  - clickhouse_internal
  - rate_limiter
  - analytics
  - export
  - import
  - yq
  - keyvalue
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
    value: 'false'
log_config:
  default_level: 5
  entry: []
  sys_log: false
nameservice_config:
  node:
  - address: ::1
    host: localhost
    node_id: 1
    port: 19001
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
static_erasure: none
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

/bin/ydbd server --node=1 --ca=${YDB_GRPC_TLS_DATA_PATH}/ca.pem --grpc-port=${GRPC_PORT} --grpcs-port=${GRPC_TLS_PORT} --yaml-config=/ydb_data/config.yaml --mon-port=8765 --ic-port=19001