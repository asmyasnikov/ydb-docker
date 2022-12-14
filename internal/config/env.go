package config

import (
	"fmt"
	"os"
	"path"
	"strconv"
)

const (
	ydbGrpcTlsDataPath = "YDB_GRPC_TLS_DATA_PATH"

	ydbDataPath = "YDB_DATA_PATH"

	ydbUseInMemoryPdisks        = "YDB_USE_IN_MEMORY_PDISKS"
	ydbUseInMemoryPdisksDefault = false

	ydbDefaultLogLevel        = "YDB_DEFAULT_LOG_LEVEL"
	ydbDefaultLogLevelDefault = 5

	grpcPort        = "GRPC_PORT"
	grpcPortDefault = 2136

	grpcTlsPort        = "GRPC_TLS_PORT"
	grpcTlsPortDefault = 2135

	monPort        = "MON_PORT"
	monPortDefault = 8765

	ydbPdiskSizeGb                = "YDB_PDISK_SIZE"
	ydbPdiskSizeGbDefault         = 80
	ydbPdiskSizeGbInMemoryDefault = 64
)

func envYdbGrpcTlsDataPath() string {
	if env, has := os.LookupEnv(ydbGrpcTlsDataPath); has {
		return env
	}
	return path.Join(ydbWorkingDir, "ydb_certs")
}

func envYdbGrpcTlsDataPathCaPem() string {
	return path.Join(envYdbGrpcTlsDataPath(), "ca.pem")
}

func envYdbGrpcTlsDataPathCertPem() string {
	return path.Join(envYdbGrpcTlsDataPath(), "cert.pem")
}

func envYdbGrpcTlsDataPathKeyPem() string {
	return path.Join(envYdbGrpcTlsDataPath(), "key.pem")
}

func envYdbDataPath() string {
	if env, has := os.LookupEnv(ydbDataPath); has {
		return env
	}
	return path.Join(ydbWorkingDir, "ydb_data")
}

func envYdbConfigPath() string {
	return path.Join(envYdbDataPath(), "config.yaml")
}

func envBindLocalStorageRequest() string {
	return path.Join(envYdbDataPath(), "bind-storage.yaml")
}

func envDefineStoragePoolsRequest() string {
	return path.Join(envYdbDataPath(), "define-storage-pools.yaml")
}

func envTenantPoolConfig() string {
	return path.Join(envYdbDataPath(), "tenant-pool.yaml")
}

func envYdbPdiskPath() string {
	if envYdbUseInMemoryPdisks() {
		return "SectorMap:1:" + strconv.Itoa(envYdbPdiskSizeGb())
	}
	return path.Join(envYdbDataPath(), "ydb.data")
}

func envYdbPdiskSizeGb() int {
	if env, has := os.LookupEnv(ydbPdiskSizeGb); has {
		if v, err := parseBytes(env); err != nil {
			panic(fmt.Errorf("cannot parse value '%s' of env '%s': %w", env, ydbPdiskSizeGb, err))
		} else {
			return int(v / GByte)
		}
	}
	if envYdbUseInMemoryPdisks() {
		return ydbPdiskSizeGbInMemoryDefault
	}
	return ydbPdiskSizeGbDefault
}

func envYdbUseInMemoryPdisks() bool {
	if env, has := os.LookupEnv(ydbUseInMemoryPdisks); has {
		b, err := strconv.ParseBool(env)
		if err != nil {
			panic(fmt.Errorf("cannot parse value '%s' of env '%s': %w", env, ydbUseInMemoryPdisks, err))
		}
		return b
	}
	return ydbUseInMemoryPdisksDefault
}

func envYdbDefaultLogLevel() int {
	if env, has := os.LookupEnv(ydbDefaultLogLevel); has {
		switch env {
		case "CRIT":
			return 2
		case "ERROR":
			return 3
		case "WARN":
			return 4
		case "NOTICE":
			return 5
		case "INFO":
			return 6
		default:
			panic(fmt.Errorf("unknown log level '%s' defined in env '%s': %w", env, ydbDefaultLogLevel))
		}
	}
	return ydbDefaultLogLevelDefault
}

func envGrpcPort() int {
	if env, has := os.LookupEnv(grpcPort); has {
		v, err := strconv.Atoi(env)
		if err != nil {
			panic(fmt.Errorf("cannot parse value '%s' of env '%s': %w", env, grpcPort, err))
		}
		return v
	}
	return grpcPortDefault
}

func envGrpcTlsPort() int {
	if env, has := os.LookupEnv(grpcTlsPort); has {
		v, err := strconv.Atoi(env)
		if err != nil {
			panic(fmt.Errorf("cannot parse value '%s' of env '%s': %w", env, grpcTlsPort, err))
		}
		return v
	}
	return grpcTlsPortDefault
}

func envMonPort() int {
	if env, has := os.LookupEnv(monPort); has {
		v, err := strconv.Atoi(env)
		if err != nil {
			panic(fmt.Errorf("cannot parse value '%s' of env '%s': %w", env, monPort, err))
		}
		return v
	}
	return monPortDefault
}
