package config

import (
	_ "embed"
	"errors"
	"os"
	"path"
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

	//go:embed templates/bind_storage_request.txt
	bindLocalStorageRequest string

	//go:embed templates/define_storage_pools_request.txt
	defineStoragePools string

	//go:embed templates/table_profile_request.txt
	tableProfilesConfig string
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
	TableProfilesConfig       string
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
}

func swapContentToFileIfNotExists(filePath string, content *string) error {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		if err = os.WriteFile(filePath, []byte(*content), 0644); err != nil {
			return err
		}
	}
	*content = filePath
	return nil
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

	if _, err := os.Stat(cfg.WorkingDir); errors.Is(err, os.ErrNotExist) {
		if err = os.MkdirAll(filepath.Dir(cfg.WorkingDir), 0777); err != nil {
			return false, err
		}
	}

	if !cfg.UseInMemoryPdisks {
		if _, err := os.Stat(cfg.Pdisk.Path); errors.Is(err, os.ErrNotExist) {
			// fallocate
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

	if err := swapContentToFileIfNotExists(path.Join(cfg.WorkingDir, "define_storage_pools_request.txt"), &cfg.DefineStoragePoolsRequest); err != nil {
		return false, err
	}

	if err := swapContentToFileIfNotExists(path.Join(cfg.WorkingDir, "bind_storage_request.txt"), &cfg.BindStorageRequest); err != nil {
		return false, err
	}

	if err := swapContentToFileIfNotExists(path.Join(cfg.WorkingDir, "table_profile_config.txt"), &cfg.TableProfilesConfig); err != nil {
		return false, err
	}

	return hasChanges, nil
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
		"STORAGE_POOL_KIND": func() string {
			return env.YdbStorePoolKind()
		},
		"STORAGE_POOL_NAME": func() string {
			return env.YdbStorePoolName()
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
	cfg.DefineStoragePoolsRequest = processTemplate(defineStoragePools)
	cfg.TableProfilesConfig = processTemplate(tableProfilesConfig)
	cfg.YdbConfig = processTemplate(config)

	return cfg, nil
}
