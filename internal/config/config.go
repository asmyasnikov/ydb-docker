package config

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/asmyasnikov/ydb-docker/internal/certs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	//go:embed config.yml
	config string
)

type Config struct {
	RamDisk bool
	Ports   struct {
		Grpc  string
		Grpcs string
		Mon   string
	}
	LogLevel string
	Data     struct {
		Path   string
		Config string
		Pdisk  struct {
			Path   string
			SizeGb string
		}
	}
	Certs *certs.Certs
}

func New(persist bool) (_ *Config, err error) {
	cfg := &Config{
		LogLevel: "5",
		Ports: struct {
			Grpc  string
			Grpcs string
			Mon   string
		}{
			Grpc:  "2136",
			Grpcs: "2135",
			Mon:   "8765",
		},
		Data: struct {
			Path   string
			Config string
			Pdisk  struct {
				Path   string
				SizeGb string
			}
		}{
			Path:   "/ydb_data/",
			Config: "/ydb_data/config.yml",
			Pdisk: struct {
				Path   string
				SizeGb string
			}{
				Path:   "/ydb_data/ydb.data",
				SizeGb: "80GB",
			},
		},
		Certs: &certs.Certs{
			Path: "/ydb_certs/",
			CA:   "/ydb_certs/ca.pem",
			Cert: "/ydb_certs/cert.pem",
			Key:  "/ydb_certs/key.pem",
		},
		RamDisk: false,
	}

	if env, has := os.LookupEnv("YDB_DATA_PATH"); has {
		cfg.Data.Path = env
		cfg.Data.Config = path.Join(env, "config.yml")
		cfg.Data.Pdisk.Path = path.Join(env, "ydb.data")
	}

	cfg.Certs, err = certs.New(persist)
	if err != nil {
		return nil, err
	}

	if _, has := os.LookupEnv("YDB_USE_IN_MEMORY_PDISKS"); has {
		cfg.RamDisk = true
	}

	if env, has := os.LookupEnv("YDB_DEFAULT_LOG_LEVEL"); has {
		switch env {
		case "CRIT":
			cfg.LogLevel = "2"
		case "ERROR":
			cfg.LogLevel = "3"
		case "WARN":
			cfg.LogLevel = "4"
		case "NOTICE":
			cfg.LogLevel = "5"
		case "INFO":
			cfg.LogLevel = "6"
		default:
			return nil, fmt.Errorf("unknow log level: %s", env)
		}
	}

	if env, has := os.LookupEnv("GRPC_PORT"); has {
		cfg.Ports.Grpc = env
	}

	if env, has := os.LookupEnv("GRPC_TLS_PORT"); has {
		cfg.Ports.Grpcs = env
	}

	if env, has := os.LookupEnv("MON_PORT"); has {
		cfg.Ports.Mon = env
	}

	if cfg.RamDisk {
		cfg.Data.Pdisk.Path = "SectorMap:1:64"
	}

	if env, has := os.LookupEnv("YDB_PDISK_SIZE"); has {
		cfg.Data.Pdisk.SizeGb = env
		if cfg.RamDisk {
			cfg.Data.Pdisk.SizeGb = "SectorMap:1:" + strings.ReplaceAll(cfg.Data.Pdisk.SizeGb, "GB", "")
		}
	}

	var buffer strings.Builder
	if err := template.Must(template.New("").Funcs(template.FuncMap{
		"YDB_PDISK_PATH": func() string {
			return cfg.Data.Pdisk.Path
		},
		"YDB_DEFAULT_LOG_LEVEL": func() string {
			return cfg.LogLevel
		},
		"GRPC_PORT": func() string {
			return cfg.Ports.Grpc
		},
		"GRPC_TLS_PORT": func() string {
			return cfg.Ports.Grpcs
		},
		"MON_PORT": func() string {
			return cfg.Ports.Mon
		},
		"YDB_PDISK_SIZE": func() string {
			return cfg.Data.Pdisk.SizeGb
		},
	}).Parse(config)).Execute(&buffer, nil); err != nil {
		return nil, err
	}

	if persist {
		if _, err := os.Stat(cfg.Data.Path); errors.Is(err, os.ErrNotExist) {
			if err = os.MkdirAll(filepath.Dir(cfg.Data.Path), 0777); err != nil {
				return nil, err
			}
		}

		if err := os.WriteFile(cfg.Data.Config, []byte(buffer.String()), 0644); err != nil {
			return nil, err
		}

	} else {
		cfg.Data.Config = buffer.String()
	}

	return cfg, nil

}
