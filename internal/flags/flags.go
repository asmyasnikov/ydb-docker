package flags

import (
	"flag"
	"fmt"
	"os"

	"github.com/asmyasnikov/ydb-docker/internal/config"
	"github.com/asmyasnikov/ydb-docker/internal/global"
)

var (
	deployCmd = flag.NewFlagSet("deploy", flag.ExitOnError)
)

func init() {
	for _, fs := range []*flag.FlagSet{deployCmd} {
		fs.StringVar(&global.YdbBinaryPath, "ydb-binary-path", "/ydbd", "Path to binary file")
		fs.StringVar(&global.YdbWorkingDir, "ydb-working-dir", "/", "Working directory for YDB cluster (the place to create directories, configuration files, disks)")
	}
}

func defaultCmd(args ...string) (*config.Config, error) {
	if err := deployCmd.Parse(args); err != nil {
		return nil, err
	}
	return config.New(config.ModeDeploy)
}

func Parse() (*config.Config, error) {
	if len(os.Args) == 1 {
		return defaultCmd()
	}

	switch os.Args[1] {
	case "deploy":
		return defaultCmd()

	default:
		return nil, fmt.Errorf("unknown subcommand '%s', see help for more details", os.Args[1])
	}
}
