package main

import (
	"context"
	"fmt"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"

	"github.com/asmyasnikov/ydb-docker/internal/flags"
	"github.com/asmyasnikov/ydb-docker/internal/log"
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
	run.Stdout, run.Stderr = log.Colored(log.NextColour())

	fmt.Fprintln(run.Stdout, run.String())

	if err = run.Start(); err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 3)

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
			cmd.Stdout, run.Stderr = log.Colored(log.NextColour())

			fmt.Fprintln(cmd.Stdout, cmd.String())

			if err = cmd.Run(); err != nil {
				panic(err)
			}
		}

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
			cmd.Stdout, run.Stderr = log.Colored(log.NextColour())

			fmt.Fprintln(cmd.Stdout, cmd.String())

			if err = cmd.Run(); err != nil {
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
			cmd.Stdout, run.Stderr = log.Colored(log.NextColour())

			fmt.Fprintln(cmd.Stdout, cmd.String())

			if err = cmd.Run(); err != nil {
				panic(err)
			}
		}
	}
	if err = run.Wait(); err != nil {
		fmt.Fprintln(run.Stderr, err.Error())
		panic(err)
	}
}
