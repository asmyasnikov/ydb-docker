package config

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/asmyasnikov/ydb-docker/internal/certs"
	"github.com/asmyasnikov/ydb-docker/internal/env"
	"github.com/asmyasnikov/ydb-docker/internal/global"
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

type Mode int

const (
	ModeUnknown = Mode(iota)
	ModeDeploy
)

type Config struct {
	Mode                      Mode
	WorkingDir                string
	BinaryPath                string
	YdbConfig                 string
	BindStorageRequest        string
	DefineStoragePoolsRequest string
	TenantPoolConfig          string
	UseInMemoryPdisks         bool
	Ports                     struct {
		Grpc  int
		Grpcs int
		Mon   int
		Ic    int
	}
	LogLevel int
	Pdisk    struct {
		Path   string
		SizeGb int
	}
	Certs *certs.Certs

	tmpFiles []string
}

func (cfg *Config) Persist(filePath string) (hasChanges bool, _ error) {
	if _, err := os.Stat(filepath.Dir(filePath)); errors.Is(err, os.ErrNotExist) {
		if err = os.MkdirAll(filepath.Dir(filePath), 0777); err != nil {
			return false, err
		}
	}

	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(filePath, []byte(cfg.YdbConfig), 0644); err != nil {
			return false, err
		}
	}
	cfg.YdbConfig = filePath

	if err := cfg.Certs.Persist(); err != nil {
		return false, err
	}

	if !cfg.UseInMemoryPdisks {
		if _, err := os.Stat(cfg.Pdisk.Path); errors.Is(err, os.ErrNotExist) {
			if err = os.MkdirAll(filepath.Dir(cfg.Pdisk.Path), 0777); err != nil {
				return false, err
			}
			pdiskFile, err := os.OpenFile(cfg.Pdisk.Path, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				return false, err
			}
			if _, err := pdiskFile.WriteAt([]byte{0}, int64(cfg.Pdisk.SizeGb)*1024*1024*1024-1); err != nil {
				return false, err
			}
			pdiskFile.Close()
			hasChanges = true
		}
	} else {
		hasChanges = true
	}

	{
		t, err := os.CreateTemp("", "tenant-pool.*.yaml")
		if err != nil {
			panic(err)
		}
		if _, err = t.WriteString(cfg.TenantPoolConfig); err != nil {
			panic(err)
		}
		t.Close()
		cfg.tmpFiles = append(cfg.tmpFiles, t.Name())
		cfg.TenantPoolConfig = t.Name()
	}

	{
		t, err := os.CreateTemp("", "define-storage-pools-request.*.yaml")
		if err != nil {
			panic(err)
		}
		if _, err = t.WriteString(cfg.DefineStoragePoolsRequest); err != nil {
			panic(err)
		}
		t.Close()
		cfg.tmpFiles = append(cfg.tmpFiles, t.Name())
		cfg.DefineStoragePoolsRequest = t.Name()
	}

	{
		t, err := os.CreateTemp("", "bind-storage-request.*.yaml")
		if err != nil {
			panic(err)
		}
		if _, err = t.WriteString(cfg.BindStorageRequest); err != nil {
			panic(err)
		}
		t.Close()
		cfg.tmpFiles = append(cfg.tmpFiles, t.Name())
		cfg.BindStorageRequest = t.Name()
	}

	return hasChanges, nil
}

func (cfg *Config) Cleanup() {
	for _, f := range cfg.tmpFiles {
		os.Remove(f)
	}
}

func New(m Mode) (*Config, error) {
	cfg := &Config{
		Mode:       m,
		WorkingDir: env.YdbDataPath(),
		BinaryPath: global.YdbBinaryPath,
		LogLevel:   env.YdbDefaultLogLevel(),
		Ports: struct {
			Grpc  int
			Grpcs int
			Mon   int
			Ic    int
		}{
			Grpc:  env.YdbGrpcPort(),
			Grpcs: env.YdbGrpcTlsPort(),
			Mon:   env.YdbMonPort(),
			Ic:    env.YdbIcPort(),
		},
		Certs: certs.New(),
		Pdisk: struct {
			Path   string
			SizeGb int
		}{
			Path:   env.YdbPdiskPath(),
			SizeGb: env.YdbPdiskSizeGb(),
		},
		UseInMemoryPdisks: env.YdbUseInMemoryPdisks(),
	}

	hostName, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	templater := template.New("").Funcs(template.FuncMap{
		"HOSTNAME": func() string {
			return hostName
		},
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
		"IC_PORT": func() int {
			return cfg.Ports.Ic
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
	})

	processTemplate := func(t string) string {
		var buffer strings.Builder
		if err := template.Must(templater.Parse(t)).Execute(&buffer, nil); err != nil {
			panic(err)
		}
		return buffer.String()
	}

	cfg.BindStorageRequest = processTemplate(bindLocalStorageRequest)
	cfg.TenantPoolConfig = processTemplate(tenantPoolConfig)
	cfg.DefineStoragePoolsRequest = processTemplate(defineStoragePools)
	cfg.YdbConfig = processTemplate(config)

	return cfg, nil
}
