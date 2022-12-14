package config

import (
	"context"
	_ "embed"
	"errors"
	"os"
	"strings"
	"text/template"
)

var (
	//go:embed templates/config.yaml
	config string

	//go:embed templates/bind-storage.request
	bindLocalStorageRequest string

	//go:embed templates/define-storage-pools.request
	defineStoragePools string

	//go:embed templates/tenant-pool.config
	tenantPoolConfig string
)

type Config struct {
	WorkingDir                string
	BinaryPath                string
	ConfigPath                string
	BindStorageRequest        string
	DefineStoragePoolsRequest string
	TenantPoolConfig          string
	UseInMemoryPdisks         bool
	Ports                     struct {
		Grpc  int
		Grpcs int
		Mon   int
	}
	LogLevel int
	Pdisk    struct {
		Path   string
		SizeGb int
	}
	Certs *Certs
}

func New(ctx context.Context, persist bool) *Config {
	cfg := &Config{
		WorkingDir:                envYdbDataPath(),
		BinaryPath:                ydbBinaryPath,
		ConfigPath:                envYdbConfigPath(),
		BindStorageRequest:        envBindLocalStorageRequest(),
		DefineStoragePoolsRequest: envDefineStoragePoolsRequest(),
		TenantPoolConfig:          envTenantPoolConfig(),
		LogLevel:                  envYdbDefaultLogLevel(),
		Ports: struct {
			Grpc  int
			Grpcs int
			Mon   int
		}{
			Grpc:  envGrpcPort(),
			Grpcs: envGrpcTlsPort(),
			Mon:   envMonPort(),
		},
		Certs: newCerts(ctx, persist),
		Pdisk: struct {
			Path   string
			SizeGb int
		}{
			Path:   envYdbPdiskPath(),
			SizeGb: envYdbPdiskSizeGb(),
		},
		UseInMemoryPdisks: envYdbUseInMemoryPdisks(),
	}

	var buffer strings.Builder
	if err := template.Must(template.New("").Funcs(template.FuncMap{
		"YDB_PDISK_PATH": func() string {
			return cfg.Pdisk.Path
		},
		"YDB_DEFAULT_LOG_LEVEL": func() int {
			return cfg.LogLevel
		},
		"GRPC_PORT": func() int {
			return cfg.Ports.Grpc
		},
		"GRPC_TLS_PORT": func() int {
			return cfg.Ports.Grpcs
		},
		"MON_PORT": func() int {
			return cfg.Ports.Mon
		},
		"YDB_PDISK_SIZE": func() int {
			return cfg.Pdisk.SizeGb
		},
		"YDB_CERTS_CA_PEM": func() string {
			return cfg.Certs.CA
		},
		"YDB_CERTS_CERT_PEM": func() string {
			return cfg.Certs.Cert
		},
		"YDB_CERTS_KEY_PEM": func() string {
			return cfg.Certs.Key
		},
	}).Parse(config)).Execute(&buffer, nil); err != nil {
		panic(err)
	}

	if !persist {
		return cfg
	}

	if _, err := os.Stat(envYdbDataPath()); errors.Is(err, os.ErrNotExist) {
		if err = os.MkdirAll(envYdbDataPath(), 0777); err != nil {
			panic(err)
		}
	}

	if err := os.WriteFile(cfg.ConfigPath, []byte(buffer.String()), 0644); err != nil {
		panic(err)
	}

	if err := os.WriteFile(cfg.BindStorageRequest, []byte(bindLocalStorageRequest), 0644); err != nil {
		panic(err)
	}

	if err := os.WriteFile(cfg.DefineStoragePoolsRequest, []byte(defineStoragePools), 0644); err != nil {
		panic(err)
	}

	if err := os.WriteFile(cfg.TenantPoolConfig, []byte(tenantPoolConfig), 0644); err != nil {
		panic(err)
	}

	return cfg
}
