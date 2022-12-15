package main

import (
	"context"
	"github.com/asmyasnikov/ydb-docker/internal/flags"
	"github.com/asmyasnikov/ydb-docker/internal/writer"
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
	run.Stdout = writer.Prefixed("[RUN] ", os.Stdout)
	run.Stderr = writer.Prefixed("[RUN] ", os.Stderr)

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
			cmd := exec.CommandContext(ctx, cfg.BinaryPath,
				"-s",
				"grpc://localhost:"+strconv.Itoa(cfg.Ports.Grpc),
				"admin",
				"blobstorage",
				"config",
				"init",
				"--yaml-file",
				cfg.YdbConfig,
			)
			cmd.Stdout = writer.Prefixed("[INIT STORAGE] ", os.Stdout)
			cmd.Stderr = writer.Prefixed("[INIT STORAGE] ", os.Stderr)

			log.Println(cmd.String())

			if err := cmd.Run(); err != nil {
				panic(err)
			}
		}

		//// register database
		//{
		//	cmd := exec.CommandContext(ctx, cfg.BinaryPath,
		//		"-s",
		//		"grpc://localhost:"+strconv.Itoa(cfg.Ports.Grpc),
		//		"admin",
		//		"database",
		//		"/local",
		//		"create",
		//		"ssd:1",
		//	)
		//	cmd.Stdout = Prefixed("[REGISTER DATABASE] ", os.Stdout)
		//	cmd.Stderr = Prefixed("[REGISTER DATABASE] ", os.Stderr)
		//
		//	log.Println(cmd.String())
		//
		//	if err := cmd.Run(); err != nil {
		//		panic(err)
		//	}
		//}

		// define storage pool
		{
			cmd := exec.CommandContext(ctx, cfg.BinaryPath,
				"-s",
				"grpc://localhost:"+strconv.Itoa(cfg.Ports.Grpc),
				"admin",
				"bs",
				"config",
				"invoke",
				"--proto-file="+cfg.DefineStoragePoolsRequest,
			)
			cmd.Stdout = writer.Prefixed("[DEFINE STORAGE POOL] ", os.Stdout)
			cmd.Stderr = writer.Prefixed("[DEFINE STORAGE POOL] ", os.Stderr)

			log.Println(cmd.String())

			if err := cmd.Run(); err != nil {
				panic(err)
			}
		}

		// init root storage
		{
			cmd := exec.CommandContext(ctx, cfg.BinaryPath,
				"-s",
				"grpc://localhost:"+strconv.Itoa(cfg.Ports.Grpc),
				"db",
				"schema",
				"execute",
				cfg.BindStorageRequest,
			)
			cmd.Stdout = writer.Prefixed("[init root storage] ", os.Stdout)
			cmd.Stderr = writer.Prefixed("[init root storage] ", os.Stderr)

			log.Println(cmd.String())

			if err := cmd.Run(); err != nil {
				panic(err)
			}
		}
	}
	if err := run.Wait(); err != nil {
		panic(err)
	}
}
