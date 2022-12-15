package main

import (
	"context"
	"fmt"
	"os"
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

	needToPrepare, err := cfg.Persist(path.Join(cfg.WorkingDir, "config.yaml"))
	if err != nil {
		panic(err)
	}

	recipe, err := os.OpenFile(path.Join(cfg.WorkingDir, "ydb_recipe.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer recipe.Close()

	_, _ = recipe.WriteString(fmt.Sprintf("======== " + time.Now().Format("2006-01-02 15:04:05") + " ========\n"))

	run := exec.CommandContext(ctx, cfg.BinaryPath,
		"server",
		"--node=1",
		"--ca="+cfg.Certs.CA,
		"--grpc-port="+strconv.Itoa(cfg.Ports.Grpc),
		"--grpcs-port="+strconv.Itoa(cfg.Ports.Grpcs),
		"--mon-port="+strconv.Itoa(cfg.Ports.Mon),
		"--ic-port="+strconv.Itoa(cfg.Ports.Ic),
		"--yaml-config="+cfg.YdbConfig,
	)

	_, _ = recipe.WriteString(run.String() + "\n")

	run.Stdout, run.Stderr = os.Stdout, os.Stderr

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

			_, _ = recipe.WriteString(cmd.String() + "\n")

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
				"blobstorage",
				"config",
				"invoke",
				"--proto-file="+cfg.DefineStoragePoolsRequest,
			)

			_, _ = recipe.WriteString(cmd.String() + "\n")

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

			_, _ = recipe.WriteString(cmd.String() + "\n")

			cmd.Stdout, run.Stderr = log.Colored(log.NextColour())

			fmt.Fprintln(cmd.Stdout, cmd.String())

			if err = cmd.Run(); err != nil {
				panic(err)
			}
		}

		// apply table profile config
		{
			cmd := exec.CommandContext(ctx, cfg.BinaryPath,
				"-s",
				"grpc://localhost:"+strconv.Itoa(cfg.Ports.Grpc),
				"admin",
				"console",
				"configs",
				"update",
				cfg.TableProfilesConfig,
			)

			_, _ = recipe.WriteString(cmd.String() + "\n")

			cmd.Stdout, run.Stderr = log.Colored(log.NextColour())

			fmt.Fprintln(cmd.Stdout, cmd.String())

			if err = cmd.Run(); err != nil {
				panic(err)
			}
		}
	}
	if err = run.Wait(); err != nil {
		panic(err)
	}
}
