package main

import (
	"context"
	"github.com/asmyasnikov/ydb-docker/internal/flags"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL)
	defer cancel()

	cfg, err := flags.Parse()
	if err != nil {
		panic(err)
	}
	defer cfg.Cleanup()

	needToPrepare, err := cfg.Persist(path.Join(cfg.WorkingDir, "config.yaml"))
	if err != nil {
		panic(err)
	}

	run := exec.CommandContext(ctx, cfg.BinaryPath,
		"server",
		"--node=1",
		"--ca="+cfg.Certs.CA,
		"--grpc-port="+strconv.Itoa(cfg.Ports.Grpc),
		"--grpcs-port="+strconv.Itoa(cfg.Ports.Grpcs),
		"--mon-port="+strconv.Itoa(cfg.Ports.Mon),
		"--ic-port="+strconv.Itoa(cfg.Ports.Ic),
		"--yaml-config="+cfg.YdbConfig,
		"--tenant-pool-file="+cfg.TenantPoolConfig,
	)
	run.Stdout = prefixed("[RUN] ", os.Stdout)
	run.Stderr = prefixed("[RUN] ", os.Stderr)

	log.Println(run.String())

	if err := run.Start(); err != nil {
		panic(err)
	}

	//if err = func(ctx context.Context) error {
	//	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	//	defer cancel()
	//	for {
	//		select {
	//		case <-ctx.Done():
	//			return ctx.Err()
	//		default:
	//			err := exec.CommandContext(ctx, "/ydb",
	//				"-e",
	//				"grpc://localhost:"+strconv.Itoa(cfg.Ports.Grpc),
	//				"-d",
	//				"/local",
	//				"scheme",
	//				"ls",
	//			)
	//			if err == nil {
	//				return nil
	//			}
	//		}
	//	}
	//}(ctx); err != nil {
	//	panic(err)
	//}

	// /ydb -e grpc://localhost:2136 -d /local scheme ls
	time.Sleep(time.Second)

	if needToPrepare {
		// initialize storage
		{
			initStorageProcess := exec.CommandContext(ctx, cfg.BinaryPath,
				"-s",
				"grpc://localhost:"+strconv.Itoa(cfg.Ports.Grpc),
				"admin",
				"blobstorage",
				"config",
				"init",
				"--yaml-file",
				cfg.YdbConfig,
			)
			initStorageProcess.Stdout = prefixed("[INIT STORAGE] ", os.Stdout)
			initStorageProcess.Stderr = prefixed("[INIT STORAGE] ", os.Stderr)

			log.Println(initStorageProcess.String())

			if err := initStorageProcess.Run(); err != nil {
				panic(err)
			}
		}

		// define storage pool
		{
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
				panic(err)
			}
		}

		// init root storage
		{
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
				panic(err)
			}
		}
	}
	if err := run.Wait(); err != nil {
		panic(err)
	}
}
