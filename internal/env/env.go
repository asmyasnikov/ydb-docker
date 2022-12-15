package env

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/asmyasnikov/ydb-docker/internal/global"
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

	icPort        = "IC_PORT"
	icPortDefault = 19001

	ydbPdiskSizeGb                = "YDB_PDISK_SIZE"
	ydbPdiskSizeGbDefault         = 80
	ydbPdiskSizeGbInMemoryDefault = 64
)

func YdbGrpcTlsDataPath() string {
	if env, has := os.LookupEnv(ydbGrpcTlsDataPath); has {
		return env
	}
	return path.Join(global.YdbWorkingDir, "ydb_certs")
}

func YdbGrpcTlsDataPathCaPem() string {
	return path.Join(YdbGrpcTlsDataPath(), "ca.pem")
}

func YdbGrpcTlsDataPathCertPem() string {
	return path.Join(YdbGrpcTlsDataPath(), "cert.pem")
}

func YdbGrpcTlsDataPathKeyPem() string {
	return path.Join(YdbGrpcTlsDataPath(), "key.pem")
}

func YdbDataPath() string {
	if env, has := os.LookupEnv(ydbDataPath); has {
		return env
	}
	return path.Join(global.YdbWorkingDir, "ydb_data")
}

func YdbConfigPath() string {
	return path.Join(YdbDataPath(), "config.yaml")
}

func YdbPdiskPath() string {
	if YdbUseInMemoryPdisks() {
		return "SectorMap:1:" + strconv.Itoa(YdbPdiskSizeGb())
	}
	return path.Join(YdbDataPath(), "ydb.data")
}

func YdbPdiskSizeGb() int {
	if env, has := os.LookupEnv(ydbPdiskSizeGb); has {
		if v, err := parseBytes(env); err != nil {
			panic(fmt.Errorf("cannot parse value '%s' of env '%s': %w", env, ydbPdiskSizeGb, err))
		} else {
			return int(v / GByte)
		}
	}
	if YdbUseInMemoryPdisks() {
		return ydbPdiskSizeGbInMemoryDefault
	}
	return ydbPdiskSizeGbDefault
}

func YdbUseInMemoryPdisks() bool {
	if env, has := os.LookupEnv(ydbUseInMemoryPdisks); has {
		b, err := strconv.ParseBool(env)
		if err != nil {
			panic(fmt.Errorf("cannot parse value '%s' of env '%s': %w", env, ydbUseInMemoryPdisks, err))
		}
		return b
	}
	return ydbUseInMemoryPdisksDefault
}

func YdbDefaultLogLevel() int {
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
			panic(fmt.Errorf("unknown log level '%s' defined in env '%s'", env, ydbDefaultLogLevel))
		}
	}
	return ydbDefaultLogLevelDefault
}

func YdbGrpcPort() int {
	if env, has := os.LookupEnv(grpcPort); has {
		v, err := strconv.Atoi(env)
		if err != nil {
			panic(fmt.Errorf("cannot parse value '%s' of env '%s': %w", env, grpcPort, err))
		}
		return v
	}
	return grpcPortDefault
}

func YdbGrpcTlsPort() int {
	if env, has := os.LookupEnv(grpcTlsPort); has {
		v, err := strconv.Atoi(env)
		if err != nil {
			panic(fmt.Errorf("cannot parse value '%s' of env '%s': %w", env, grpcTlsPort, err))
		}
		return v
	}
	return grpcTlsPortDefault
}

func YdbMonPort() int {
	if env, has := os.LookupEnv(monPort); has {
		v, err := strconv.Atoi(env)
		if err != nil {
			panic(fmt.Errorf("cannot parse value '%s' of env '%s': %w", env, monPort, err))
		}
		return v
	}
	return monPortDefault
}

func YdbIcPort() int {
	if env, has := os.LookupEnv(icPort); has {
		v, err := strconv.Atoi(env)
		if err != nil {
			panic(fmt.Errorf("cannot parse value '%s' of env '%s': %w", env, icPort, err))
		}
		return v
	}
	return icPortDefault
}
