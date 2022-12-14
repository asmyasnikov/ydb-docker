package config

import "flag"

var (
	ydbBinaryPath string
	ydbWorkingDir string
)

func init() {
	flag.StringVar(&ydbBinaryPath, "ydb-binary-path", "/ydbd", "Path to binary file")
	flag.StringVar(&ydbWorkingDir, "ydb-working-dir", "/", "Working directory for YDB cluster (the place to create directories, configuration files, disks)")

	flag.Parse()
}
