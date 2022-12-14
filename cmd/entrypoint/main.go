package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/asmyasnikov/ydb-docker/internal/config"
)

func prepare(ctx context.Context, cfg *config.Config) error {
	// start storage process
	storageProcess := exec.CommandContext(ctx, cfg.BinaryPath,
		"server",
		"--node=1",
		"--ca="+cfg.Certs.CA,
		"--grpc-port="+strconv.Itoa(cfg.Ports.Grpc),
		"--grpcs-port="+strconv.Itoa(cfg.Ports.Grpcs),
		"--mon-port="+strconv.Itoa(cfg.Ports.Mon),
		"--ic-port=19001",
		"--yaml-config="+cfg.ConfigPath,
	)
	storageProcess.Stdout = prefixed("[storage] ", os.Stdout)
	storageProcess.Stderr = prefixed("[storage] ", os.Stderr)

	log.Println(storageProcess.String())

	if err := storageProcess.Start(); err != nil {
		return err
	}
	defer func() {
		_ = storageProcess.Process.Kill()
	}()

	time.Sleep(3 * time.Second)

	// initialize storage
	initStorageProcess := exec.CommandContext(ctx, cfg.BinaryPath,
		"-s",
		"grpc://localhost:"+strconv.Itoa(cfg.Ports.Grpc),
		"admin",
		"blobstorage",
		"config",
		"init",
		"--yaml-file",
		cfg.ConfigPath,
	)
	initStorageProcess.Stdout = prefixed("[init storage] ", os.Stdout)
	initStorageProcess.Stderr = prefixed("[init storage] ", os.Stderr)

	log.Println(initStorageProcess.String())

	if err := initStorageProcess.Run(); err != nil {
		return err
	}

	// register database
	registerDatabaseProcess := exec.CommandContext(ctx, cfg.BinaryPath,
		"-s",
		"grpc://localhost:"+strconv.Itoa(cfg.Ports.Grpc),
		"admin",
		"database",
		"/local",
		"create",
		"ssd:1",
	)
	registerDatabaseProcess.Stdout = prefixed("[register database] ", os.Stdout)
	registerDatabaseProcess.Stderr = prefixed("[register database] ", os.Stderr)

	log.Println(registerDatabaseProcess.String())

	if err := registerDatabaseProcess.Run(); err != nil {
		return err
	}

	// define storage pool
	defineStoragePoolProcess := exec.CommandContext(ctx, cfg.BinaryPath,
		"-s",
		"grpc://localhost:"+strconv.Itoa(cfg.Ports.Grpc),
		"admin",
		"bs",
		"config",
		"invoke",
		"--proto-file="+cfg.DefineStoragePoolsRequest,
	)
	defineStoragePoolProcess.Stdout = prefixed("[define storage pool] ", os.Stdout)
	defineStoragePoolProcess.Stderr = prefixed("[define storage pool] ", os.Stderr)

	log.Println(defineStoragePoolProcess.String())

	if err := defineStoragePoolProcess.Run(); err != nil {
		return err
	}

	// init root storage
	initRootStorageProcess := exec.CommandContext(ctx, cfg.BinaryPath,
		"-s",
		"grpc://localhost:"+strconv.Itoa(cfg.Ports.Grpc),
		"db",
		"schema",
		"execute",
		cfg.BindStorageRequest,
	)
	initRootStorageProcess.Stdout = prefixed("[init root storage] ", os.Stdout)
	initRootStorageProcess.Stderr = prefixed("[init root storage] ", os.Stderr)

	log.Println(initRootStorageProcess.String())

	if err := initRootStorageProcess.Run(); err != nil {
		return err
	}

	// stop storage process
	if err := storageProcess.Process.Kill(); err != nil {
		return err
	}

	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL)
	defer cancel()

	cfg := config.New(ctx, true)

	if !cfg.UseInMemoryPdisks {
		if _, err := os.Stat(cfg.Pdisk.Path); errors.Is(err, os.ErrNotExist) {
			if err = os.MkdirAll(filepath.Dir(cfg.Pdisk.Path), 0777); err != nil {
				panic(err)
			}
			pdiskFile, err := os.OpenFile(cfg.Pdisk.Path, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				panic(err)
			}
			if _, err := pdiskFile.WriteAt([]byte{0}, int64(cfg.Pdisk.SizeGb)*1024*1024*1024-1); err != nil {
				panic(err)
			}
			pdiskFile.Close()

			if err := prepare(ctx, cfg); err != nil {
				panic(err)
			}
		}
	} else {
		if err := prepare(ctx, cfg); err != nil {
			panic(err)
		}
	}

	// run compute with storage at single process
	run := exec.CommandContext(ctx, cfg.BinaryPath,
		"server",
		"--node=1",
		"--ca="+cfg.Certs.CA,
		"--grpc-port="+strconv.Itoa(cfg.Ports.Grpc),
		"--grpcs-port="+strconv.Itoa(cfg.Ports.Grpcs),
		"--mon-port="+strconv.Itoa(cfg.Ports.Mon),
		"--yaml-config="+cfg.ConfigPath,
		"--tenant-pool-file="+cfg.TenantPoolConfig,
	)
	run.Stdout = prefixed("[run] ", os.Stdout)
	run.Stderr = prefixed("[run] ", os.Stderr)

	log.Println(run.String())

	if err := run.Run(); err != nil {
		panic(err)
	}
}
